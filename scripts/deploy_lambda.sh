#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${ROOT_DIR}/.build/lambda"
ZIP_PATH="${BUILD_DIR}/function.zip"
TRUST_POLICY_PATH="${BUILD_DIR}/trust-policy.json"

AWS_REGION="${AWS_REGION:-eu-north-1}"
PROFILE_NAME="${AWS_PROFILE:-}"
FUNCTION_NAME="${FUNCTION_NAME:-cicd-microservice}"
ROLE_NAME="${ROLE_NAME:-${FUNCTION_NAME}-exec-role}"
ROLE_ARN="${ROLE_ARN:-}"
RUNTIME="${RUNTIME:-provided.al2023}"
ARCH="${ARCH:-arm64}"
MEMORY_SIZE="${MEMORY_SIZE:-256}"
TIMEOUT="${TIMEOUT:-10}"
DEPLOY_MODE="${DEPLOY_MODE:-full}"

if [[ "${DEPLOY_MODE}" != "full" && "${DEPLOY_MODE}" != "update-only" ]]; then
  echo "DEPLOY_MODE must be 'full' or 'update-only'"
  exit 1
fi

if ! command -v aws >/dev/null 2>&1; then
  echo "aws CLI is not installed"
  exit 1
fi

if ! command -v zip >/dev/null 2>&1; then
  echo "zip is not installed"
  exit 1
fi

# An explicitly empty AWS_PROFILE/AWS_DEFAULT_PROFILE breaks AWS CLI with profile "".
if [[ -z "${PROFILE_NAME}" ]]; then
  unset AWS_PROFILE AWS_DEFAULT_PROFILE || true
fi

AWS_CMD=(aws --region "${AWS_REGION}")
if [[ -n "${PROFILE_NAME}" ]]; then
  AWS_CMD+=(--profile "${PROFILE_NAME}")
fi

echo "Using AWS region: ${AWS_REGION}"
if [[ -n "${PROFILE_NAME}" ]]; then
  echo "Using AWS profile: ${PROFILE_NAME}"
else
  echo "Using AWS profile: default credential chain"
fi
echo "Deploy mode: ${DEPLOY_MODE}"
echo "Function name: ${FUNCTION_NAME}"

mkdir -p "${BUILD_DIR}"

ACCOUNT_ID="$("${AWS_CMD[@]}" sts get-caller-identity --query Account --output text)"
if [[ -z "${ROLE_ARN}" ]]; then
  ROLE_ARN="arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"
fi

if [[ "${DEPLOY_MODE}" == "full" ]] && ! "${AWS_CMD[@]}" iam get-role --role-name "${ROLE_NAME}" >/dev/null 2>&1; then
  cat > "${TRUST_POLICY_PATH}" <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

  echo "Creating IAM role ${ROLE_NAME}..."
  "${AWS_CMD[@]}" iam create-role \
    --role-name "${ROLE_NAME}" \
    --assume-role-policy-document "file://${TRUST_POLICY_PATH}" >/dev/null

  "${AWS_CMD[@]}" iam attach-role-policy \
    --role-name "${ROLE_NAME}" \
    --policy-arn "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole" >/dev/null

  # IAM propagation.
  sleep 10
else
  if [[ "${DEPLOY_MODE}" == "full" ]]; then
    echo "IAM role ${ROLE_NAME} already exists."
  fi
fi

echo "Building Lambda bootstrap..."
(
  cd "${ROOT_DIR}"
  GOOS=linux GOARCH="${ARCH}" CGO_ENABLED=0 go build -tags lambda.norpc -o "${BUILD_DIR}/bootstrap" ./cmd/lambda
)

echo "Packaging function.zip..."
(
  cd "${BUILD_DIR}"
  rm -f "${ZIP_PATH}"
  zip -q -j "${ZIP_PATH}" bootstrap
)

if "${AWS_CMD[@]}" lambda get-function --function-name "${FUNCTION_NAME}" >/dev/null 2>&1; then
  echo "Updating existing Lambda function..."
  "${AWS_CMD[@]}" lambda update-function-code \
    --function-name "${FUNCTION_NAME}" \
    --zip-file "fileb://${ZIP_PATH}" >/dev/null

  if [[ "${DEPLOY_MODE}" == "full" ]]; then
    "${AWS_CMD[@]}" lambda update-function-configuration \
      --function-name "${FUNCTION_NAME}" \
      --role "${ROLE_ARN}" \
      --runtime "${RUNTIME}" \
      --handler bootstrap \
      --memory-size "${MEMORY_SIZE}" \
      --timeout "${TIMEOUT}" \
      --architectures "${ARCH}" >/dev/null
  fi
else
  if [[ "${DEPLOY_MODE}" == "update-only" ]]; then
    echo "Lambda function ${FUNCTION_NAME} does not exist and DEPLOY_MODE=update-only."
    echo "Create it once with DEPLOY_MODE=full, then use update-only in CI."
    exit 1
  fi

  echo "Creating Lambda function..."
  "${AWS_CMD[@]}" lambda create-function \
    --function-name "${FUNCTION_NAME}" \
    --runtime "${RUNTIME}" \
    --handler bootstrap \
    --role "${ROLE_ARN}" \
    --memory-size "${MEMORY_SIZE}" \
    --timeout "${TIMEOUT}" \
    --architectures "${ARCH}" \
    --zip-file "fileb://${ZIP_PATH}" >/dev/null
fi

echo "Waiting for function to become active..."
"${AWS_CMD[@]}" lambda wait function-active-v2 --function-name "${FUNCTION_NAME}"

if "${AWS_CMD[@]}" lambda get-function-url-config --function-name "${FUNCTION_NAME}" >/dev/null 2>&1; then
  echo "Updating existing Function URL config..."
  "${AWS_CMD[@]}" lambda update-function-url-config \
    --function-name "${FUNCTION_NAME}" \
    --auth-type NONE >/dev/null
else
  echo "Creating Function URL config..."
  "${AWS_CMD[@]}" lambda create-function-url-config \
    --function-name "${FUNCTION_NAME}" \
    --auth-type NONE >/dev/null
fi

add_permission_if_needed() {
  local statement_id="$1"
  shift
  set +e
  output="$("${AWS_CMD[@]}" lambda add-permission "$@" --statement-id "${statement_id}" 2>&1)"
  exit_code=$?
  set -e
  if [[ ${exit_code} -ne 0 ]]; then
    if echo "${output}" | grep -q "ResourceConflictException"; then
      echo "Permission ${statement_id} already exists."
      return 0
    fi
    echo "${output}"
    return ${exit_code}
  fi
}

add_permission_if_needed "${FUNCTION_NAME}-url-invoke-url" \
  --function-name "${FUNCTION_NAME}" \
  --action "lambda:InvokeFunctionUrl" \
  --principal "*" \
  --function-url-auth-type NONE

add_permission_if_needed "${FUNCTION_NAME}-url-invoke-fn" \
  --function-name "${FUNCTION_NAME}" \
  --action "lambda:InvokeFunction" \
  --principal "*" \
  --invoked-via-function-url

FUNCTION_URL="$("${AWS_CMD[@]}" lambda get-function-url-config --function-name "${FUNCTION_NAME}" --query FunctionUrl --output text)"

echo ""
echo "Lambda deployment completed."
echo "Function: ${FUNCTION_NAME}"
echo "URL: ${FUNCTION_URL}"

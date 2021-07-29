#!/usr/bin/env bash

set -e
set -o pipefail
set -o nounset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1


INSTANCES=()
UPDATE_INSTANCES=()

SERVICES=()
for s in "${SERVICES[@]}"; do
    create_service "${s}" small "${s}-$$" &
    INSTANCES+=("${s}-$$")
done

UPDATE_SERVICES=("csb-azure-mysql" "csb-azure-mssql" "csb-azure-mssql-failover-group" "csb-azure-postgresql")
for s in "${UPDATE_SERVICES[@]}"; do
    if [ "${s}" == "csb-azure-mssql-failover-group" ]; then
        plan="small-v2"
    elif [ "${s}" == "csb-azure-mssql" ]; then
        plan="small-v2"
    else
        plan="small"
    fi
    create_service "${s}" "$plan" "${s}-$$" &
    INSTANCES+=("${s}-$$")
    UPDATE_INSTANCES+=("${s}-$$")
done

INSTANCES+=()

NO_TLS_SERVICES=("csb-azure-mysql" "csb-azure-postgresql")

for s in "${NO_TLS_SERVICES[@]}"; do
    create_service "${s}" small "${s}-no-tls-$$" "{\"use_tls\":false}" &
    INSTANCES+=("${s}-no-tls-$$")
done

if wait; then
    RESULT=0
    for s in "${INSTANCES[@]}"; do
        if [ $# -gt 0 ]; then
            if "${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}"; then
                echo "SUCCEEDED: ${SCRIPT_DIR}/../cf-validate-credhub.sh ${s}"
            else
                RESULT=1
                echo "FAILED: ${SCRIPT_DIR}/../cf-validate-credhub.sh" "${s}"
                break
            fi
        fi

        TEST_CMD="${SCRIPT_DIR}/../cf-run-spring-music-test.sh ${s}"

        if in_list ${s} "${UPDATE_INSTANCES[@]}"; then
            echo "Will run cf update-service test on ${s}"
            TEST_CMD="${SCRIPT_DIR}/../cf-run-spring-music-test.sh ${s} medium"
        fi

        if ${TEST_CMD}; then
            echo "SUCCEEDED: ${TEST_CMD}"
        else
            RESULT=1
            echo "FAILED: ${TEST_CMD}"
            break
        fi
    done
else
    echo "FAILED creating one or more service instances"
fi

for s in "${INSTANCES[@]}"; do
    delete_service "${s}" &
done

wait

if [ ${RESULT} -eq 0 ]; then
    echo "SUCCEEDED: $0"
else
    echo "FAILED: $0"
fi

exit ${RESULT}

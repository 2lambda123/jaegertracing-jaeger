#!/bin/bash

set -euf -o pipefail

print_help() {
  echo "Usage: $0 [-b binary]"
  echo "-b: Which binary to build: 'all-in-one' (default) or 'jaeger' (v2)"
  echo "-h: Print help"
  exit 1
}

BINARY='all-in-one'
compose_file=docker-compose/monitor/docker-compose.yml

while getopts "b:h" opt; do
	case "${opt}" in
	b)
		BINARY=${OPTARG}
		;;
	?)
		print_help
		;;
	esac
done

if [ "$BINARY" == "jaeger" ]; then
  compose_file=docker-compose/monitor/docker-compose-v2.yml
fi

timeout=300
end_time=$((SECONDS + timeout))
success="false"

check_service_health() {
  local service_name=$1
  local url=$2
  echo "Checking health of service: $service_name at $url"

  local wait_seconds=10
  local curl_params=(
    --silent
    --output
    /dev/null
    --write-out
    "%{http_code}"
  )
  while [ $SECONDS -lt $end_time ]; do
    if [[ "$(curl "${curl_params[@]}" "${url}")" == "200" ]]; then
      echo "✅ $service_name is healthy"
      return 0
    fi
    echo "Waiting for $service_name to be healthy..."
    sleep $wait_seconds
  done

  echo "❌ ERROR: $service_name did not become healthy in time"
  return 1
}

# Function to check if all services are healthy
wait_for_services() {
  echo "Waiting for services to be up and running..."
  check_service_health "Jaeger" "http://localhost:16686"
  check_service_health "Prometheus" "http://localhost:9090/graph"
  # Grafana is not actually important for the functional test,
  # but it at least validates that the docker-compose file is correct.
  check_service_health "Grafana" "http://localhost:3000"
}

# Function to validate the service metrics
validate_service_metrics() {
    local service=$1
    # Time constants in milliseconds
    local fiveMinutes=300000
    local oneMinute=60000
    local fifteenSec=15000 # Prometheus is also configured to scrape every 15sec.
    # When endTs=(blank) the server will default it to now().
    local url="http://localhost:16686/api/metrics/calls?service=${service}&endTs=&lookback=${fiveMinutes}&step=${fifteenSec}&ratePer=${oneMinute}"
    response=$(curl -s "$url")
    service_name=$(echo "$response" | jq -r 'if .metrics and .metrics[0] then .metrics[0].labels[] | select(.name=="service_name") | .value else empty end')
    if [ "$service_name" != "$service" ]; then
      echo "⏳ No metrics found for service '$service'"
      return 1
    fi
    # Store the values in an array
    mapfile -t metric_points < <(echo "$response" | jq -r '.metrics[0].metricPoints[].gaugeValue.doubleValue')
    echo "Metric datapoints found for service '$service': " "${metric_points[@]}"
    # Check that all values are non-zero
    local non_zero_count=0
    for value in "${metric_points[@]}"; do
      if [[ $(echo "$value > 0.0" | bc) == "1" ]]; then
        non_zero_count=$((non_zero_count + 1))
      else
        echo "❌ ERROR: Zero values not expected"
        return 1
      fi
    done
    if [ $non_zero_count -lt 3 ]; then
      echo "⏳ Expecting at least 3 non-zero data points"
      return 1
    fi
    return 0
}

check_spm() {
  local wait_seconds=10
  local successful_service=0
  services_list=("driver" "customer" "mysql" "redis" "frontend" "route" "ui")
  for service in "${services_list[@]}"; do
    echo "Processing service: $service"
    while [ $SECONDS -lt $end_time ]; do
      if validate_service_metrics "$service"; then
        echo "✅ Found all expected metrics for service '$service'"
        successful_service=$((successful_service + 1))
        break
      fi
      sleep $wait_seconds
    done
  done
  if [ $successful_service -lt ${#services_list[@]} ]; then
    echo "❌ ERROR: Expected metrics from ${#services_list[@]} services, found only ${successful_service}"
    exit 1
  else
    echo "✅ All services metrics are returned by the API"
  fi
}

dump_logs() {
  echo "::group:: docker logs"
  docker compose -f $compose_file logs
  echo "::endgroup::"
}

teardown_services() {
  if [[ "$success" == "false" ]]; then
    dump_logs
  fi
  docker compose -f $compose_file down
}

main() {
  (cd docker-compose/monitor && make build BINARY="$BINARY")
  if [ "$BINARY" == "jaeger" ]; then
    (make dev-v2 DOCKER_COMPOSE_ARGS="-d")
  else
    (make dev DOCKER_COMPOSE_ARGS="-d")
  fi
  wait_for_services
  check_spm
  success="true"
}

trap teardown_services EXIT INT

# Run the main function
main

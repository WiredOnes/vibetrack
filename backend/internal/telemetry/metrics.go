package telemetry

type gaugeMetric string

var ()

type counterMetric string

var (
	HTTPRequestsTotal counterMetric = "http_requests_total"
	GRPCCallsTotal    counterMetric = "grpc_calls_total"
	ExecutedJobsTotal counterMetric = "executed_jobs_total"
	DBRequestsTotal   counterMetric = "db_requests_total"
)

type histogramMetric string

var ()

type summaryMetric string

var (
	HTTPRequestDurationSeconds summaryMetric = "http_request_duration_seconds"
	GRPCCallDurationSeconds    summaryMetric = "grpc_call_duration_seconds"
	JobDurationSeconds         summaryMetric = "job_duration_seconds"
)

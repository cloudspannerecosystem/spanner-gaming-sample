apiVersion: v1
kind: ConfigMap
metadata:
  name: spanner-config
data:
  SPANNER_PROJECT_ID: ${project_id}
  SPANNER_INSTANCE_ID: ${instance_id}
  SPANNER_DATABASE_ID: ${database_id}

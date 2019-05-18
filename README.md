# ofc-bot

OpenFaaS Cloud admin-bot for Slack

## Commands:

### List all users

```
/users
```

### List all user-functions

```
/functions
```

### List all functions for a given user

```
/functions <username>
```

### Get metrics for a function

```
/metrics <function>
```

This will show the success / error count for the last `24h` window.

### Get logs for a function

```
/logs <function>
```

View the last 100 lines of logs for the first pod found in the cluster with a matching function name.

## Other config

Template: `golang-middleware`

Secret name: `USERNAME-ofc-bot-secrets`

Secrets literals:

* `basic-auth-password` - for your gateway admin user
* `token` - Slack token for verification

Seal your own secrets:

```sh
export SCM_USER="alexellis"
export BASIC_AUTH=""
export SLACK_TOKEN=""
export PAYLOAD_SECRET=""

faas-cli cloud seal \
  --name $SCM_USER-ofc-bot-secrets \
  --literal basic-auth-password=$BASIC_AUTH \
  --literal token=$SLACK_TOKEN \
  --literal payload-secret=$PAYLOAD_SECRET
```

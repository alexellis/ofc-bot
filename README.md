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

## Other config

Template: `golang-middleware`

Secret name: `USERNAME-ofc-bot-secrets`

Secrets literals:

* `basic-auth-password` - for your gateway admin user
* `token` - Slack token for verification

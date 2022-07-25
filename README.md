# Prometheus Matrix bot

A simple [Matrix](https://matrix.org) bot for Prometheus alerting

TODO: improve documentatio and cleanup

### Setup

- Create a Matrix user for the bot and acquire the access token for it
- Set up the bot (Docker preferred) with the environment variables defined in `docker-compose.yml` (note: the bot will accept room invites only from the defined admin user)

### Alertmanager configuration

You should set up receivers with webhook configs with name beginning with "matrix-" followed by the room ID to post to, for example:

```
receivers:
- name: matrix-!dfgh5yhdF54yFTJH:example.com
  webhook_configs:
    - url: http://127.0.0.1:8080/alert
      send_resolved: true
```

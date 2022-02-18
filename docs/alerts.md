# Alerts

The following table lists the currently available alerts in Prometheus/Alert Manager with their short description.

**Note:** In order to enable the alerts you'll need to uncomment everything in **"monitoring/prometheus/rules.yaml"** and to update the 
sensitive data in **"monitoring/alertmanager/config.yml"**.

| Name                             | Description                                                         |
|----------------------------------|---------------------------------------------------------------------|
| `LowValidatorsParticipationRate` | Alerting if the participation rate is under 66.66 % (2/3)           |
| `LowFeeAccountAmount`            | Alerting if the Fee Account Amount is under recommended value.      |
| `LowOperatorAccountAmount`       | Alerting if the Operator Account Amount is under recommended value. |
                                                                                   
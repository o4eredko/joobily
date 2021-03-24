# Joobily
Slack bot that can be configured to answer questions. People in our company actively used it to get Wi-Fi password, information about vacation, coronavirus etc.
## Exapmle of usage
![usage](./attachments/screencast.gif)
## Running
You have to enter SLACK_BOT_ID, SLACK_ACCESS_TOKEN, SLACK_VERIFICATION_TOKEN, SLACK_SIGNING_SECRET variables in .env file before running this app.

```docker-compose up```
## Tools
This app is using ElasticSearch to make routing of questions more flexible.
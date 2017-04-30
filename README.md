mastodon-notifications-sqs

* Still temporary placement *

# Function
Monitor mastodon notification Streaming and push to AWS SQS.
Multiple instances are supported.

# setting.yml
Place setting.yml in the same hierarchy.
Multiple serverConfs can be specified in the list.

```yml
awsRegion: [AWS Region]
queueURL: [SQS endpoint URL]

serverConfs:
  - serverName: [name of instance - for display at notification]
    serverURL: [Instance URL - no trailing slash]
    clientID: [Mustdon client ID - described later]
    clientSecret: [Mastodon client secret key - described later]
    account: [Mastodon account name]
    password: [Mastodon password]
```

# Preparation for using Mastodon API

Requests the mastdon server to issue a client ID / secret.
Call curl as follows.
Please modify the contents as appropriate.

```bash
curl -X POST -sS https://xxxxxxxxxxxxxx/api/v1/apps \
   -F "client_name=xxxxxxxxxx" \
   -F "redirect_uris=urn:ietf:wg:oauth:2.0:oob" \
   -F "scopes=read write follow"
```


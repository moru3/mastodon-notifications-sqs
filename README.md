mastodon-notifications-sqs

*まだ仮置き*

# 機能
マストドン通知Streamingを監視しAWS SQSへプッシュします。
複数インスタンス対応しています。

# setting.yml
同階層にsetting.ymlを配置します。
serverConfsはリストで複数指定できます。

```yml
awsRegion: [AWSリージョン]
queueURL: [SQSエンドポイントURL]

serverConfs:
  - serverName: [インスタンスの名称(通知時の表示用)]
    serverURL: [インスタンスのURL(末尾スラッシュなし)]
    clientID: [マストドンクライアントID(後述)]
    clientSecret: [マストドンクライアントシークレットキー(後述)]
    account: [マストドンユーザ名]
    password: [マストドンパスワード]
```

# マストドンAPI利用の下準備

マストドンサーバに対してクライアントID/シークレットを発行を要求します。
curlだと以下のように呼び出します。
内容は適宜修正してください。

```bash
curl -X POST -sS https://xxxxxxxxxxxxxx/api/v1/apps \
   -F "client_name=xxxxxxxxxx" \
   -F "redirect_uris=urn:ietf:wg:oauth:2.0:oob" \
   -F "scopes=read write follow"
```


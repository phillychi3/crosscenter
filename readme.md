# CrossCenter

> [!WARNING]
> No guarantee of account security after use (Twitter)<br>
> Under development; unstable or catastrophic accidents may occur

## Tutorials

### Initial Setup

1. Rename `set.yaml.example` to `set.yaml`
2. change PostText<br> Example: `"{text} #sometag author: {author} url:{url} Date:{date}"`<br>keywords: author, text, url, date

### Threads Setup

1. Open https://developers.facebook.com
2. Create an app for Threads
3. Go to App roles -> Roles
   ![Facebook App Role](docs/image/fb-app-role.png)
4. Add your account
5. Go to https://www.threads.net/settings/account and view Website permissions setting (PC only) -> invite and accept
   ![Threads Invite](docs/image/fb-app-threads-invite.png)
6. Open Graph API
   ![Facebook Graph API](docs/image/fb-app-graph.png)
7. Change `.facebook.com` to `threads.net`
   ![Facebook Graph API Change](docs/image/fb-app-graph-api.png)
8. Copy **Access Token** to `set.yaml` under `Access_Token`
   ![Facebook Access Token](docs/image/fb-app-graph-api-token.png)
9. Go to App settings -> Basic and copy App secret to `set.yaml` under `Client_Secret`
   ![Facebook App ID](docs/image/fb-app-id.png)
10. Open Use cases and customize Access the Threads API, Add `threads_content_publish` and `threads_basic`
    ![Facebook Use Cases](docs/image/fb-app-usecases.png)

### Twitter Setup

1. Go to https://developer.x.com/en and login
2. Create a project
3. Copy API `Key` and `Secret` to `CONSUMER_KEY` and `CONSUMER_SECRET` respectively
   Copy Access `Token` and `Secret` to `ACCESS_TOKEN` and `ACCESS_TOKEN_SECRET` respectively
   ![Twitter App Keys](docs/image/twitter-app-keys.png)
4. **Warning**: The next steps involve risks. If you don't want to use a live account, stop here. If continuing, consider creating a 2FA-free account.
5. Open developer tools (F12) and go to the Network tab, search for `https://x.com/home`
6. Copy `auth_token` and `ct0` to `Auth_token` and `Ct0` respectively
7. Set `REAC_ACCOUNT_MODE` to `true`

## Supported Sites

### RSS

- Read

### Twitter

- Read
- Post

### Threads

- Read
- Post

## TODO:

- [x] Discord webhook
- [ ] Instagram
- [ ] Facebook

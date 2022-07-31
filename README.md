# cd-tilde
one shot VPN

Creates OpenVPN server, which sends configuration file to your telegram when ready to connect.

0. Create selectel account
1. Install terraform from https://www.terraform.io/downloads
2. Create your own file terraform.tfvars in current directory that contains [selectel credentials](https://kb.selectel.com/docs/cloud/servers/tools/how_to_use_terraform/#provider-configuration) and [telegram token](https://core.telegram.org/bots#6-botfather)
```terraform
selectel-account   = "012345" # number of selected account (mentioned in control panel )
project-id         = "0123456789abcdef0123456789abcdef" # selectel project id (shows after creating project in cloud platform)
openstack-user     = "user" # credentials of cloud platform project for openstack
openstack-pass     = "pA55w0rd"
selectel-api-token = "AbCdEfGhIjKlMnOpQrStUvWxYz012345" # API Key (creates in Setings of Control Panel)
telegram-bot-token = "123456789:AbCdEfGhIjKlMnOpQrStUvWxYz012345678" # bot token 
```
3. Run `terraform init` to initialize terraform providers
4. Generate or copy existing certificate in bot/ folder. Please note, that you should use fullchain certificate for the embedded web server.
5. Specify your telegram bot token, your public domain for which has generated HTTPS certificate, filenames of certificate keypair, public available port and telegram id of bot admin in environment variables.
```bash
export BOT_PORT=8443
export BOT_DOMAIN=example.com
export BOT_APITOKEN=123456789:AbCdEfGhIjKlMnOpQrStUvWxYz012345678
export BOT_KEY=privkey.pem
export BOT_CERT=fullchain.pem
export BOT_OWNER=012345678
```
6. Specify the ids of telegram users which allowed to use this bot in bot/config.json file in the following format:
"user_id": allowed_duration_in_seconds. default value 600 seconds if the value is negative. When a new user who was not specified in config.json use /start command of this bot, the bot admin receives a message with first name, username and user id of a new one. You can add it by hand or ignore the message.
```json
{
    "234567890": 3600,
    "345678901": -1
}
```
7. Run `go run .` from bot/ directory

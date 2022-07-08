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
4. Specify your telegram bot token and your public domain which has generated HTTPS certificate in bot/bot.go
```bash
sed -i "s/YOUR_SECRET_TOKEN/123456789:AbCdEfGhIjKlMnOpQrStUvWxYz012345678/;s/your-public-domain/example.com/" bot/bot.go
```
5. Generate or copy existing certificate in bot/ folder. Please note, that you should use fullchain certificate for the embedded web server.
8. Run `go run .` from bot/ directory

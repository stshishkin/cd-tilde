# cd-tilde
one shot VPN

Creates OpenVPN server, which sends configuration file to your telegram when ready to connect.

0. Create selectel account
1. Install terraform from https://www.terraform.io/downloads
2. Creates your own file terraform.tfvars in current directory that contains [selectel credentials](https://kb.selectel.com/docs/cloud/servers/tools/how_to_use_terraform/#provider-configuration) and [telegram token](https://core.telegram.org/bots#6-botfather)
```terraform
selectel-account   = "012345" # number of selected account (mentioned in control panel )
project-id         = "0123456789abcdef0123456789abcdef" # selectel project id (shows after creating project in cloud platform)
openstack-user     = "user" # credentials of cloud platform project for openstack
openstack-pass     = "pA55w0rd"
selectel-api-token = "AbCdEfGhIjKlMnOpQrStUvWxYz012345" # API Key (creates in Setings of Control Panel)
telegram-bot-token = "123456789:AbCdEfGhIjKlMnOpQrStUvWxYz012345678" # bot token 
telegram-chat      = "100123456"
```
3. run `terraform init` to initialize terraform providers
4. run `terraform apply` create everything
5. enjoy your new VPN that costs 1,36 ₽/hour
6. run `terraform destroy` when no longer needed

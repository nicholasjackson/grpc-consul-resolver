data_dir= "/tmp/consul"

connect {
  enabled = true
  proxy {
   allow_managed_api_registration = true
   allow_managed_root = true
  }
}

retry_interval = "1s"

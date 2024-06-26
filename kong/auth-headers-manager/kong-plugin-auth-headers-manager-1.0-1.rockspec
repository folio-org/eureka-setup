package = "kong-plugin-auth-headers-manager"
version = "1.0-1"
local pluginName = "auth-headers-manager"

description = {
  summary = "A Kong plugin that will convert cookies into headers",
  license = "Apache 2.0"
}

source = {
  url = "locally generated",
  tag = "v1.0-1"
}

dependencies = {
  "lua ~> 5.1"
}

build = {
  type = "builtin",
  modules = {
    ["kong.plugins.auth-headers-manager.handler"] = "handler.lua",
    ["kong.plugins.auth-headers-manager.schema"] = "schema.lua"
  }
}

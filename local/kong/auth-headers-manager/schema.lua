local typedefs = require "kong.db.schema.typedefs"

return {
  name = "auth-headers-manager",
  fields = {
    { consumer = typedefs.no_consumer },
    { protocols = typedefs.protocols_http },
    { config = {
      type = "record",
      fields = {
        { set_okapi_header = {
          type = "boolean",
          default = true,
          required = false },
        },
        { set_authorization_header = {
          type = "boolean",
          default = false,
          required = false },
        },
      },
    } }
  }
}

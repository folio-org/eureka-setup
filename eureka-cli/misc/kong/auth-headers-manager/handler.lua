local AuthTokenManager = {
  PRIORITY = 1010,
  VERSION = "1.0-1",
}

local kong = kong
local okapiTokenHeader = "X-Okapi-Token"
local authorizationHeader = 'Authorization'
local folioAccessTokenCookie = "folioAccessToken"

local function getCookies()
  local cookieHeader = kong.request.get_header("cookie")
  if not cookieHeader then
    return {}
  end

  local cookies = {}
  local iterator, err = ngx.re.gmatch(cookieHeader, "([^\\s]+)=([^\\s]+);\\s*", "io")
  if not iterator or err then
    return {}
  end

  while true do
    local m, error = iterator()
    if error then
      return {}
    end

    if not m then
      break
    end

    cookies[m[1]] = m[2]
  end

  return cookies
end

local function startsWith(str, start)
  return str:sub(1, #start) == start
end

local function getAccessTokenFromHeaders()
  local request = kong.request
  local okapiAuthToken = request.get_header(okapiTokenHeader)
  local authorizationToken = request.get_header(authorizationHeader)

  if authorizationToken then
    if not startsWith(authorizationToken, "Bearer ") then
      return kong.response.error(404,
        'Invalid authorization header, value must start with Bearer',
        { ["Content-Type"] = "application/json" })
    end

    local authorizationTokenValue = authorizationToken:sub(8, authorizationToken:len());
    if okapiAuthToken and okapiAuthToken ~= authorizationTokenValue then
      return kong.response.error(404,
        'X-Okapi-Token is not equal to Authorization token',
        { ["Content-Type"] = "application/json" })
    end

    return { source = authorizationHeader, token = authorizationTokenValue }
  end

  return { source = okapiTokenHeader, token = okapiAuthToken }
end

local function getAccessToken()
  local cookies = getCookies()
  local folioAccessToken = cookies[folioAccessTokenCookie]
  local headersToken = getAccessTokenFromHeaders();
  if folioAccessToken then
    if headersToken.token and headersToken.token ~= folioAccessToken then
      return kong.response.error(404,
        headersToken.source .. ' token is not equal to ' .. folioAccessTokenCookie .. ' token in cookies',
        { ["Content-Type"] = "application/json" })
    end

    return { source = folioAccessTokenCookie, token = folioAccessToken }
  end

  return headersToken
end

function AuthTokenManager:access(conf)
  local accessToken = getAccessToken()
  if not accessToken then
    return
  end

  kong.log.debug('is Okapi token enabled: ', conf.set_okapi_header)
  if conf.set_okapi_header then
    if accessToken.source == folioAccessTokenCookie and not kong.request.get_header(okapiTokenHeader) then
      kong.log.debug("Setting X-Okapi-Token header from cookie value...", accessToken.token)
      kong.service.request.set_header(okapiTokenHeader, accessToken.token)
    end
  end

  kong.log.debug('is Authorization token enabled: ', conf.set_authorization_header)
  if conf.set_authorization_header then
    if accessToken.source == folioAccessTokenCookie and not kong.request.get_header(authorizationHeader) then
      kong.log.debug("Setting Authorization header from cookie value...")
      kong.service.request.set_header(authorizationHeader, "Bearer " .. accessToken.token)
    end
  end
end

return AuthTokenManager

{{ $cfg := .Config }}

local ngx_balancer = require("ngx.balancer")
local resolver = require("resty.dns.resolver")
local dns_cache = ngx.shared.dns_cache
local hostnames = {}

local _M = {}

local function resolve_dns(hostname)
    local r, err = resolver:new{
        nameservers = {"127.0.0.11"},
        retrans = 5,
        timeout = 2000,
    }

    if not r then
        ngx.log(ngx.ERR, "failed to instantiate the resolver: ", err)
        return
    end

    local answers, err = r:query(hostname, nil, {})
    if not answers then
        ngx.log(ngx.ERR, "failed to query the DNS server for ", hostname, ": ", err)
        return
    end

    if answers.errcode then
        ngx.log(ngx.ERR, "server returned error code for ", hostname, ": ", answers.errcode, ": ", answers.errstr)
        return
    end

    ngx.log(ngx.INFO, "DNS resolution for ", hostname, " returned ", #answers, " answers")

    for _, ans in ipairs(answers) do
        if ans.address then
            local ok, err = dns_cache:set(hostname, ans.address, 45)
            if not ok then
                ngx.log(ngx.ERR, "Failed to set ", hostname, " in cache: ", err)
                return
            end

            ngx.log(ngx.INFO, "Resolved ", hostname, " to ", ans.address)
            return ans.address
        end
    end
end

local function resolve_all_dns()
    ngx.log(ngx.INFO, "Resolving DNS for ", #hostnames, " hostnames")

    for _, hostname in ipairs(hostnames) do
        resolve_dns(hostname)
    end
end

local function contains_hostname(value)
  for _, v in ipairs(hostnames) do
      if v == value then
          return true
      end
  end
  return false
end

function _M.init_worker()
    local ok, err = ngx.timer.every(60, resolve_all_dns)
    if not ok then
        ngx.log(ngx.ERR, "failed to create the timer: ", err)
    end
end

function _M.balance()
    local hostname = ngx.ctx.upstream_host
    if not hostname then
        ngx.log(ngx.ERR, "Missing 'upstream_host' variable")
        return ngx.exit(500)
    end

    local port = ngx.ctx.upstream_port
    if not port then
        ngx.log(ngx.ERR, "Missing 'upstream_port' variable")
        return ngx.exit(500)
    end

    local ip = dns_cache:get(hostname)
    if not ip then
        ngx.log(ngx.ERR, "DNS resolution for ", hostname, " not found in cache")
        return ngx.exit(500)
    end

    ngx_balancer.set_more_tries(1)

    local ok, err = ngx_balancer.set_current_peer(ip, port)
    if not ok then
        ngx.log(ngx.ERR, "Failed to set the current peer: ", err)
        return ngx.exit(500)
    end
end

function _M.set_peer(host, port)
    ngx.ctx.upstream_host = host
    ngx.ctx.upstream_port = port

    if not contains_hostname(host) then
        table.insert(hostnames, host)
    end

    local ip = dns_cache:get(host)
    if not ip then
        resolve_dns(host)
    end
end

return _M
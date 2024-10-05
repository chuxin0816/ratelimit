local key = KEYS[1]                 -- Redis key，令牌桶的唯一标识
local rate = tonumber(ARGV[1])      -- 每秒产生的令牌数
local capacity = tonumber(ARGV[2])  -- 桶的最大容量
local now = tonumber(ARGV[3])       -- 当前时间戳（毫秒）
local requested = tonumber(ARGV[4]) -- 请求的令牌数量

-- 获取当前的令牌数量和上次更新时间
local last_tokens = tonumber(redis.call("GET", key .. ":tokens"))
local last_refreshed = tonumber(redis.call("GET", key .. ":timestamp"))

-- 如果没有数据，则初始化令牌桶
if last_tokens == nil or last_refreshed == nil then
    last_tokens = capacity
    last_refreshed = now
end

-- 计算自上次刷新后的新令牌数量
local delta = (now - last_refreshed) * rate / 1000  -- 改为精确到毫秒
local new_tokens = math.min(capacity, last_tokens + delta)

-- 判断请求的令牌数量是否足够
local allowed = 0
if new_tokens >= requested then
    -- 扣除请求的令牌并允许通过
    new_tokens = new_tokens - requested
    allowed = 1
end

-- 更新令牌桶的状态
redis.call("SET", key .. ":tokens", new_tokens)
redis.call("SET", key .. ":timestamp", now)

-- 返回是否允许通过
return allowed
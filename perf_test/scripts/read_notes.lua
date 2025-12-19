
local min_id = 1
local max_id = 117852

math.randomseed(os.time())

function request()
    local note_id = math.random(min_id, max_id)
    local path = "/api/notes/" .. note_id
    return wrk.format("GET", path)
end

function response(status, headers, body)
    if status ~= 200 then
        if status ~= 404 then
            io.write(string.format("Error: status=%d, body=%s\n", status, body))
        end
    end
end

function done(summary, latency, requests)
    io.write("------------------------------\n")
    io.write("Результаты чтения заметок:\n")
    io.write("------------------------------\n")
    io.write(string.format("Всего запросов: %d\n", summary.requests))
    io.write(string.format("Ошибок: %d\n", summary.errors.status))
    io.write(string.format("Таймаутов: %d\n", summary.errors.timeout))
    io.write(string.format("Средняя latency: %.2f ms\n", latency.mean / 1000))
    io.write(string.format("Max latency: %.2f ms\n", latency.max / 1000))
    io.write(string.format("Stdev latency: %.2f ms\n", latency.stdev / 1000))
    io.write(string.format("P50 latency: %.2f ms\n", latency:percentile(50) / 1000))
    io.write(string.format("P90 latency: %.2f ms\n", latency:percentile(90) / 1000))
    io.write(string.format("P99 latency: %.2f ms\n", latency:percentile(99) / 1000))
    io.write(string.format("RPS: %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write("------------------------------\n")
end
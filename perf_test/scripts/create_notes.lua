wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"

local counter = 0

function request()
    counter = counter + 1

    return wrk.format(nil, "/api/notes", nil, "{}")
end

function response(status, headers, body)
    if status ~= 200 and status ~= 201 then
        io.write(string.format("Error: status=%d, body=%s\n", status, body))
    end
end

function done(summary, latency, requests)
    io.write("Результаты создания заметок:\n")
    io.write(string.format("Всего запросов: %d\n", summary.requests))
    io.write(string.format("Ошибок: %d\n", summary.errors.status))
    io.write(string.format("Таймаутов: %d\n", summary.errors.timeout))
    io.write(string.format("Средняя latency: %.2f ms\n", latency.mean / 1000))
    io.write(string.format("Max latency: %.2f ms\n", latency.max / 1000))
    io.write(string.format("Stdev latency: %.2f ms\n", latency.stdev / 1000))
    io.write(string.format("RPS: %.2f\n", summary.requests / (summary.duration / 1000000)))
end
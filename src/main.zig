const std = @import("std");
const keccak = @import("keccak");
const secp = @import("secp256k1");

const Mode = enum { normal, vanity, exclude };
            if (matches(digest[12..], &shared.config)) {
    pattern: []const u8 = "",
    prefix: bool = true,
                var address: [40]u8 = undefined;
                encodeHex(digest[12..], &address);
    excluded: [256]bool = [_]bool{false} ** 256,
};

const Arguments = struct {
    count: ?usize = null,
    workers: ?usize = null,
};

const Shared = struct {
    output: std.io.BufferedWriter(64 * 1024, std.fs.File.Writer),
    write_mutex: std.Thread.Mutex = .{},
    next_index: std.atomic.Value(usize) = std.atomic.Value(usize).init(0),
    completed: std.atomic.Value(usize) = std.atomic.Value(usize).init(0),
    attempts: std.atomic.Value(u64) = std.atomic.Value(u64).init(0),
    failed: std.atomic.Value(bool) = std.atomic.Value(bool).init(false),
    total: usize,
    config: Config,
};

pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    const args = try std.process.argsAlloc(allocator);
    const arguments = try parseArguments(args);
    const total = arguments.count orelse try promptCount(allocator);
    const config = try promptConfig(allocator);
    const cpu_count = std.Thread.getCpuCount() catch 1;
    const default_workers = cpu_count;
    const worker_count = @min(arguments.workers orelse default_workers, total);

    const file = try std.fs.cwd().createFile("wallets.csv", .{ .truncate = true });
    defer file.close();
    try file.writeAll("address,private_key\n");
    var shared = Shared{
        .output = .{ .unbuffered_writer = file.writer() },
        .total = total,
        .config = config,
    };

    var pool: std.Thread.Pool = undefined;
    try pool.init(.{ .allocator = allocator, .n_jobs = @intCast(worker_count) });
    defer pool.deinit();
    var wait_group: std.Thread.WaitGroup = .{};
    for (0..worker_count) |_| pool.spawnWg(&wait_group, worker, .{&shared});

    while (shared.completed.load(.acquire) < total and !shared.failed.load(.acquire)) {
        drawProgress(shared.completed.load(.acquire), total, shared.attempts.load(.monotonic), worker_count);
        std.time.sleep(100 * std.time.ns_per_ms);
    }
    wait_group.wait();
    shared.output.flush() catch {
        shared.failed.store(true, .release);
    };
    drawProgress(total, total, shared.attempts.load(.monotonic), worker_count);
    std.debug.print("\nGenerated {d} wallets using {d} CPU workers.\n", .{ total, worker_count });
    if (shared.failed.load(.acquire)) return error.GenerationFailed;
}

fn worker(shared: *Shared) void {
    var generator = secp.Generator.init() catch {
        shared.failed.store(true, .release);
        return;
    };
    defer generator.deinit();
    var local_attempts: u64 = 0;
    defer if (local_attempts != 0) {
        _ = shared.attempts.fetchAdd(local_attempts, .monotonic);
    };

    while (!shared.failed.load(.acquire)) {
        const index = shared.next_index.fetchAdd(1, .monotonic);
        if (index >= shared.total) break;

        var private_key: [32]u8 = undefined;
        var public_key: [65]u8 = undefined;
        while (true) {
            std.crypto.random.bytes(&private_key);
            local_attempts += 1;
            if (local_attempts & 1023 == 0) {
                _ = shared.attempts.fetchAdd(local_attempts, .monotonic);
                local_attempts = 0;
            }
            public_key = generator.publicKey(&private_key) catch continue;
            const digest = keccak.hash(public_key[1..]);
            var address: [40]u8 = undefined;
            encodeHex(digest[12..], &address);
            if (matches(&address, &shared.config)) {
                shared.write_mutex.lock();
                defer shared.write_mutex.unlock();
                var line: [108]u8 = undefined;
                @memcpy(line[0..40], &address);
                line[40] = ',';
                line[41] = '0';
                line[42] = 'x';
                encodeHex(&private_key, line[43..107]);
                line[107] = '\n';
                shared.output.writer().writeAll(&line) catch {
                    shared.failed.store(true, .release);
                    return;
                };
                _ = shared.completed.fetchAdd(1, .release);
                break;
            }
        }
    }
}

fn matches(address_bytes: []const u8, config: *const Config) bool {
    if (config.mode == .normal) return true;
    if (config.mode == .vanity) {
        const start = if (config.prefix) 0 else 40 - config.pattern.len;
        for (config.pattern, 0..) |character, index| {
            const byte = address_bytes[(start + index) / 2];
            const nibble = if ((start + index) & 1 == 0) byte >> 4 else byte & 0x0f;
            if (nibble != hexValue(character)) return false;
        }
        return true;
    }
    for (address_bytes) |byte| {
        if (config.excluded[hexDigits[byte >> 4]] or config.excluded[hexDigits[byte & 0x0f]]) return false;
    }
    return true;
}

const hexDigits = "0123456789abcdef";

fn hexValue(character: u8) u8 {
    return if (character <= '9') character - '0' else character - 'a' + 10;
}

fn encodeHex(bytes: []const u8, output: []u8) void {
    for (bytes, 0..) |byte, index| {
        output[index * 2] = hexDigits[byte >> 4];
        output[index * 2 + 1] = hexDigits[byte & 0x0f];
    }
}

fn promptConfig(allocator: std.mem.Allocator) !Config {
    var config = Config{ .mode = .normal };
    std.debug.print("\nWallet generation mode\n  1. Generate normal wallets\n  2. Generate vanity wallets\n  3. Generate wallets excluding characters\n\nSelect an option (1-3): ", .{});
    const choice = try readLine(allocator);
    if (std.mem.eql(u8, choice, "1")) return config;
    if (std.mem.eql(u8, choice, "2")) {
        config.mode = .vanity;
        std.debug.print("Match location\n  1. After 0x (address start)\n  2. At the address end\n\nSelect a location (1-2): ", .{});
        const location = try readLine(allocator);
        if (!std.mem.eql(u8, location, "1") and !std.mem.eql(u8, location, "2")) return error.InvalidLocation;
        config.prefix = std.mem.eql(u8, location, "1");
        std.debug.print("Enter a vanity pattern, up to 10 hexadecimal characters: ", .{});
        const pattern = try readLine(allocator);
        if (pattern.len == 0 or pattern.len > 10) return error.InvalidPattern;
        for (pattern) |*character| {
            if (!std.ascii.isHex(character.*)) return error.InvalidPattern;
            character.* = std.ascii.toLower(character.*);
        }
        config.pattern = pattern;
        return config;
    }
    if (std.mem.eql(u8, choice, "3")) {
        config.mode = .exclude;
        std.debug.print("Enter hexadecimal characters to exclude, separated by commas (example: b,6,f): ", .{});
        const input = try readLine(allocator);
        const values = std.mem.trimRight(u8, input, ",");
        if (values.len == 0) return error.InvalidExcludedCharacter;
        var iterator = std.mem.splitScalar(u8, values, ',');
        while (iterator.next()) |part| {
            const character = std.mem.trim(u8, part, " \t");
            if (character.len != 1 or !std.ascii.isHex(character[0])) return error.InvalidExcludedCharacter;
            config.excluded[std.ascii.toLower(character[0])] = true;
        }
        return config;
    }
    return error.InvalidMode;
}

fn readLine(allocator: std.mem.Allocator) ![]u8 {
    const input = try std.io.getStdIn().reader().readUntilDelimiterAlloc(allocator, '\n', 128);
    const trimmed = std.mem.trim(u8, input, " \r\n");
    std.mem.copyForwards(u8, input[0..trimmed.len], trimmed);
    return input[0..trimmed.len];
}

fn promptCount(allocator: std.mem.Allocator) !usize {
    std.debug.print("Number of wallets to generate: ", .{});
    const input = try readLine(allocator);
    const count = try std.fmt.parseUnsigned(usize, input, 10);
    if (count == 0) return error.InvalidCount;
    return count;
}

fn drawProgress(completed: usize, total: usize, attempts: u64, workers: usize) void {
    const width: usize = 32;
    const filled = if (total == 0) width else (completed * width) / total;
    std.debug.print("\r\x1b[32m[", .{});
    for (0..filled) |_| std.debug.print("#", .{});
    std.debug.print("\x1b[31m", .{});
    for (filled..width) |_| std.debug.print("-", .{});
    const percent = if (total == 0) 100 else (completed * 100) / total;
    std.debug.print("\x1b[0m] {d:>3}% | {d}/{d} | tries {d} | {d} CPU", .{ percent, completed, total, attempts, workers });
}

fn parseArguments(args: []const []const u8) !Arguments {
    var result = Arguments{};
    var index: usize = 1;
    while (index < args.len) : (index += 2) {
        if (index + 1 >= args.len) return error.InvalidArguments;
        const value = try std.fmt.parseUnsigned(usize, args[index + 1], 10);
        if (value == 0) return error.InvalidArguments;
        if (std.mem.eql(u8, args[index], "--count")) {
            result.count = value;
        } else if (std.mem.eql(u8, args[index], "--worker")) {
            result.workers = value;
        } else {
            return error.InvalidArguments;
        }
    }
    return result;
}
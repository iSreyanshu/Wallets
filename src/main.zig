const std = @import("std");
const keccak = @import("keccak");
const secp = @import("secp256k1");

const Mode = enum { normal, vanity, exclude };

const Config = struct {
    mode: Mode,
    pattern: []const u8 = "",
    prefix: bool = true,
    excluded: [256]bool = [_]bool{false} ** 256,
};

const Arguments = struct {
    count: ?usize = null,
    workers: ?usize = null,
};

const Shared = struct {
    file: std.fs.File,
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
    const default_workers = std.math.mul(usize, cpu_count, 100) catch std.math.maxInt(usize);
    const worker_count = @min(arguments.workers orelse default_workers, total);

    const file = try std.fs.cwd().createFile("wallets.csv", .{ .truncate = true });
    defer file.close();
    try file.writeAll("address,private_key\n");
    var shared = Shared{ .file = file, .total = total, .config = config };

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

    while (!shared.failed.load(.acquire)) {
        const index = shared.next_index.fetchAdd(1, .monotonic);
        if (index >= shared.total) break;

        var private_key: [32]u8 = undefined;
        var public_key: [65]u8 = undefined;
        while (true) {
            std.crypto.random.bytes(&private_key);
            _ = shared.attempts.fetchAdd(1, .monotonic);
            public_key = generator.publicKey(&private_key) catch |err| switch (err) {
                error.InvalidPrivateKey => continue,
                else => {
                    shared.failed.store(true, .release);
                    return;
                },
            };
            const digest = keccak.hash(public_key[1..]);
            if (matches(digest[12..], shared.config)) {
                shared.write_mutex.lock();
                defer shared.write_mutex.unlock();
                shared.file.writer().print("{f},0x{f}\n", .{
                    std.fmt.fmtSliceHexLower(digest[12..]),
                    std.fmt.fmtSliceHexLower(&private_key),
                }) catch {
                    shared.failed.store(true, .release);
                    return;
                };
                _ = shared.completed.fetchAdd(1, .release);
                break;
            }
        }
    }
}

fn matches(address: []const u8, config: Config) bool {
    return switch (config.mode) {
        .normal => true,
        .vanity => if (config.prefix) std.mem.startsWith(u8, address, config.pattern) else std.mem.endsWith(u8, address, config.pattern),
        .exclude => for (address) |digit| {
            if (config.excluded[digit]) break false;
        } else true,
    };
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
        var iterator = std.mem.splitScalar(u8, input, ',');
        while (iterator.next()) |part| {
            if (part.len != 1 or !std.ascii.isHex(part[0])) return error.InvalidExcludedCharacter;
            config.excluded[std.ascii.toLower(part[0])] = true;
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
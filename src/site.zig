const std = @import("std");

pub fn main() !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    const args = try std.process.argsAlloc(allocator);
    const output_path = try outputPath(args);
    const html = try std.fs.cwd().readFileAlloc(allocator, "README.md", 1024 * 1024 * 4);
    const directory = std.fs.path.dirname(output_path) orelse ".";

    try std.fs.cwd().makePath(directory);
    const output = try std.fs.cwd().createFile(output_path, .{ .truncate = true });
    defer output.close();
    try output.writeAll(html);
}

fn outputPath(args: []const []const u8) ![]const u8 {
    if (args.len == 1) return ".github/index.html";
    if (args.len == 3 and std.mem.eql(u8, args[1], "--output") and args[2].len > 0) return args[2];
    return error.InvalidArguments;
}
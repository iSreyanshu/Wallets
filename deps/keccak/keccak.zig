const std = @import("std");

const rc = [_]u64{
    1, 0x8082, 0x800000000000808a, 0x8000000080008000,
    0x808b, 0x80000001, 0x8000000080008081, 0x8000000000008009,
    0x8a, 0x88, 0x80008009, 0x8000000a, 0x8000808b, 0x800000000000008b,
    0x8000000000008089, 0x8000000000008003, 0x8000000000008002,
    0x8000000000000080, 0x800a, 0x800000008000000a,
    0x8000000080008081, 0x8000000000008080, 0x80000001, 0x8000000080008008,
};

const rotation = [5][5]u6{
    .{ 0, 36, 3, 41, 18 }, .{ 1, 44, 10, 45, 2 },
    .{ 62, 6, 43, 15, 61 }, .{ 28, 55, 25, 21, 56 },
    .{ 27, 20, 39, 8, 14 },
};

fn permute(state: *[25]u64) void {
    for (rc) |round_constant| {
        var column: [5]u64 = undefined;
        for (0..5) |x| column[x] = state[x] ^ state[x + 5] ^ state[x + 10] ^ state[x + 15] ^ state[x + 20];
        for (0..5) |x| {
            const correction = column[(x + 4) % 5] ^ std.math.rotl(u64, column[(x + 1) % 5], 1);
            for (0..5) |y| state[x + 5 * y] ^= correction;
        }

        var rotated: [25]u64 = undefined;
        for (0..5) |x| {
            for (0..5) |y| {
                rotated[y + 5 * ((2 * x + 3 * y) % 5)] = std.math.rotl(u64, state[x + 5 * y], rotation[x][y]);
            }
        }
        for (0..5) |x| {
            for (0..5) |y| {
                state[x + 5 * y] = rotated[x + 5 * y] ^ ((~rotated[(x + 1) % 5 + 5 * y]) & rotated[(x + 2) % 5 + 5 * y]);
            }
        }
        state[0] ^= round_constant;
    }
}

pub fn hash(input: []const u8) [32]u8 {
    var state = [_]u64{0} ** 25;
    const rate = 136;
    var offset: usize = 0;
    while (offset + rate <= input.len) : (offset += rate) {
        for (0..rate / 8) |i| state[i] ^= std.mem.readInt(u64, input[offset + i * 8 ..][0..8], .little);
        permute(&state);
    }
    var block = [_]u8{0} ** rate;
    const remaining = input[offset..];
    @memcpy(block[0..remaining.len], remaining);
    block[remaining.len] = 0x01;
    block[rate - 1] |= 0x80;
    for (0..rate / 8) |i| state[i] ^= std.mem.readInt(u64, block[i * 8 ..][0..8], .little);
    permute(&state);

    var output: [32]u8 = undefined;
    for (0..4) |i| std.mem.writeInt(u64, output[i * 8 ..][0..8], state[i], .little);
    return output;
}
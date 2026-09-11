const std = @import("std");
const Curve = std.crypto.ecc.Secp256k1;

pub const Generator = struct {
    pub fn init() !Generator {
        return .{};
    }

    pub fn deinit(_: *Generator) void {}

    pub fn publicKey(self: *Generator, private_key: *const [32]u8) ![65]u8 {
        _ = self;
        if (std.mem.allEqual(u8, private_key, 0)) return error.InvalidPrivateKey;
        Curve.scalar.rejectNonCanonical(private_key.*, .big) catch return error.InvalidPrivateKey;
        const point = Curve.basePoint.mul(private_key.*, .big) catch return error.InvalidPrivateKey;
        return point.toUncompressedSec1();
    }
};

test "generator returns the secp256k1 base point for private key one" {
    var generator = try Generator.init();
    defer generator.deinit();
    var private_key = [_]u8{0} ** 32;
    private_key[31] = 1;
    const public_key = try generator.publicKey(&private_key);
    try std.testing.expectEqual(@as(u8, 4), public_key[0]);
    try std.testing.expectEqualStrings(
        "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" ++
            "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
        &hexString(public_key[1..]),
    );
}

fn hexString(bytes: []const u8) [128]u8 {
    var output: [128]u8 = undefined;
    const digits = "0123456789abcdef";
    for (bytes, 0..) |byte, index| {
        output[index * 2] = digits[byte >> 4];
        output[index * 2 + 1] = digits[byte & 0x0f];
    }
    return output;
}
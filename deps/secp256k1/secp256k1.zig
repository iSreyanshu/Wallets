const c = @cImport({
    @cInclude("openssl/ec.h");
    @cInclude("openssl/obj_mac.h");
    @cInclude("openssl/bn.h");
});

pub const Generator = struct {
    group: *c.EC_GROUP,
    context: *c.BN_CTX,
    order: *c.BIGNUM,
    private_bn: *c.BIGNUM,
    public_point: *c.EC_POINT,

    pub fn init() !Generator {
        const group = c.EC_GROUP_new_by_curve_name(c.NID_secp256k1) orelse return error.OpenSSL;
        const context = c.BN_CTX_new() orelse return error.OpenSSL;
        const order = c.BN_new() orelse return error.OpenSSL;
        const private_bn = c.BN_new() orelse return error.OpenSSL;
        const public_point = c.EC_POINT_new(group) orelse return error.OpenSSL;
        if (c.EC_GROUP_get_order(group, order, context) != 1) return error.OpenSSL;
        return .{ .group = group, .context = context, .order = order, .private_bn = private_bn, .public_point = public_point };
    }

    pub fn deinit(self: *Generator) void {
        c.EC_POINT_free(self.public_point);
        c.BN_free(self.private_bn);
        c.BN_free(self.order);
        c.BN_CTX_free(self.context);
        c.EC_GROUP_free(self.group);
    }

    pub fn publicKey(self: *Generator, private_key: *const [32]u8) ![65]u8 {
        if (c.BN_bin2bn(private_key, 32, self.private_bn) == null) return error.OpenSSL;
        if (c.BN_is_zero(self.private_bn) == 1 or c.BN_cmp(self.private_bn, self.order) >= 0) return error.InvalidPrivateKey;
        if (c.EC_POINT_mul(self.group, self.public_point, self.private_bn, null, null, self.context) != 1) return error.OpenSSL;
        var public_key: [65]u8 = undefined;
        const length = c.EC_POINT_point2oct(self.group, self.public_point, c.POINT_CONVERSION_UNCOMPRESSED, &public_key, public_key.len, self.context);
        if (length != public_key.len) return error.OpenSSL;
        return public_key;
    }
};
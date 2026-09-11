const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const exe = b.addExecutable(.{
        .name = "evm",
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });
    exe.root_module.addImport("keccak", b.createModule(.{
        .root_source_file = b.path("deps/keccak/keccak.zig"),
        .target = target,
        .optimize = optimize,
    }));
    exe.root_module.addImport("secp256k1", b.createModule(.{
        .root_source_file = b.path("deps/secp256k1/secp256k1.zig"),
        .target = target,
        .optimize = optimize,
    }));
    exe.linkLibC();
    exe.linkSystemLibrary("crypto");
    b.installArtifact(exe);

    const run = b.addRunArtifact(exe);
    run.step.dependOn(b.getInstallStep());
    if (b.args) |args| run.addArgs(args);

    const run_step = b.step("run", "Generate evm wallets");
    run_step.dependOn(&run.step);
}

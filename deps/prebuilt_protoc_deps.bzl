"""
GENERATED FILE - DO NOT EDIT (created via @build_stack_rules_proto//cmd/depsgen)
"""

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

VERSION = "3.20.1"

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)

def prebuilt_protoc_deps():
    """prebuilt_protoc dependency macro
    """
    prebuilt_protoc_linux_amd64()  # via <TOP>
    prebuilt_protoc_linux_arm64()  # via <TOP>
    prebuilt_protoc_osx_amd64()  # via <TOP>
    prebuilt_protoc_osx_arm64()  # via <TOP>
    prebuilt_protoc_windows()  # via <TOP>

def prebuilt_protoc_linux_amd64():
    _maybe(
        http_archive,
        name = "prebuilt_protoc_linux_amd64",
        sha256 = "3a0e900f9556fbcac4c3a913a00d07680f0fdf6b990a341462d822247b265562",
        urls = [
            "https://github.com/google/protobuf/releases/download/v%s/protoc-%s-linux-x86_64.zip" % (VERSION, VERSION),
        ],
        build_file_content = """
filegroup(
    name = "protoc",
    srcs = ["bin/protoc"],
    visibility = ["//visibility:public"],
)
""",
    )

def prebuilt_protoc_linux_arm64():
    _maybe(
        http_archive,
        name = "prebuilt_protoc_linux_arm64",
        sha256 = "8a5a51876259f934cd2acc2bc59dba0e9a51bd631a5c37a4b9081d6e4dbc7591",
        urls = [
            "https://github.com/google/protobuf/releases/download/v%s/protoc-%s-linux-aarch_64.zip" % (VERSION, VERSION),
        ],
        build_file_content = """
filegroup(
    name = "protoc",
    srcs = ["bin/protoc"],
    visibility = ["//visibility:public"],
)
""",
    )

def prebuilt_protoc_osx_amd64():
    _maybe(
        http_archive,
        name = "prebuilt_protoc_osx_amd64",
        sha256 = "b4f36b18202d54d343a66eebc9f8ae60809a2a96cc2d1b378137550bbe4cf33c",
        urls = [
            "https://github.com/google/protobuf/releases/download/v%s/protoc-%s-osx-x86_64.zip" % (VERSION, VERSION),
        ],
        build_file_content = """
filegroup(
    name = "protoc",
    srcs = ["bin/protoc"],
    visibility = ["//visibility:public"],
)
""",
    )

def prebuilt_protoc_osx_arm64():
    _maybe(
        http_archive,
        name = "prebuilt_protoc_osx_arm64",
        sha256 = "b362acae78542872bb6aac8dba73aaf0dc6e94991b8b0a065d6c3e703fec2a8b",
        urls = [
            "https://github.com/google/protobuf/releases/download/v%s/protoc-%s-osx-aarch_64.zip" % (VERSION, VERSION),
        ],
        build_file_content = """
filegroup(
    name = "protoc",
    srcs = ["bin/protoc"],
    visibility = ["//visibility:public"],
)
""",
    )

def prebuilt_protoc_windows():
    _maybe(
        http_archive,
        name = "prebuilt_protoc_windows",
        sha256 = "2291c634777242f3bf4891b082cebc6dd495ae621fbf751b27e800b83369a345",
        urls = [
            "https://github.com/google/protobuf/releases/download/v%s/protoc-%s-win32.zip" % (VERSION, VERSION),
        ],
        build_file_content = """
filegroup(
    name = "protoc",
    srcs = ["bin/protoc.exe"],
    visibility = ["//visibility:public"],
)
""",
    )

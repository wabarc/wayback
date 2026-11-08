{ pkgs }: {
    deps = [
        pkgs.go
        pkgs.tor
        pkgs.wget
        pkgs.chromium
        pkgs.ffmpeg
        pkgs.libwebp
        pkgs.youtube-dl
        pkgs.you-get
        pkgs.ipfs
    ];
}

#!/bin/bash

set -eu pipefail

export GPG_TTY=$(tty)

WAYBACK_SIGNING_KEY="${WAYBACK_SIGNING_KEY:-}"
WAYBACK_SIGNING_PASSPHARSE="${WAYBACK_SIGNING_PASSPHARSE:-}"
VERSION="${VERSION:-1.0}"
WORKDIR="/rpmbuild"

cat > ~/.rpmmacros<< EOF
%_topdir /rpmbuild
%_signature gpg
%_gpg_name Wayback Archiver
EOF

mkdir -p "${WORKDIR}/{BUILD,RPMS,SOURCES,SPECS,SRPMS}"

gpg --import --yes --pinentry-mode loopback --passphrase "${WAYBACK_SIGNING_PASSPHARSE}" <<< "${WAYBACK_SIGNING_KEY}"

rpmbuild -bb --define "_wayback_version ${VERSION}" "${WORKDIR}/SPECS/wayback.spec"

find "${WORKDIR}/RPMS/x86_64" -type f -name "*.rpm" -exec rpm --verbose --define "_gpg_sign_cmd_extra_args --pinentry-mode loopback --passphrase ${WAYBACK_SIGNING_PASSPHARSE}" --addsign {} \;

find "${WORKDIR}/RPMS/x86_64" -type f -name "*.rpm" -exec rpm -qpi {} \;


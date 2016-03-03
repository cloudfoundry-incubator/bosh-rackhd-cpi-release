#!/usr/bin/env bash

set -e -x


source bosh-cpi-release/ci/tasks/utils.sh

check_param S3_ACCESS_KEY_ID
check_param S3_SECRET_ACCESS_KEY

# Creates an integer version number from the semantic version format
# May be changed when we decide to fully use semantic versions for releases
integer_version=`cut -d "." -f1 release-version-semver/number`
mkdir -p promote
echo $integer_version > promote/integer_version

cp -r bosh-cpi-release promote/bosh-cpi-release
cd promote/bosh-cpi-release

#source /etc/profile.d/chruby.sh
#chruby 2.1.2

set +x
echo creating config/private.yml with blobstore secrets
cat > config/private.yml << EOF
---
blobstore:
  s3:
    bucket_name: bosh-rackhd-cpi-blobs
    access_key_id: ${S3_ACCESS_KEY_ID}
    secret_access_key: ${S3_SECRET_ACCESS_KEY}
EOF
set -x

echo "using bosh CLI version..."
bosh version

echo "finalizing CPI release..."
echo '' | bosh create release --force --with-tarball --version $integer_version
bosh finalize release dev_releases/bosh-rackhd-cpi/*.tgz --version $integer_version

rm config/private.yml

git diff | cat
git add .

git config --global user.email emccmd-eng@emc.com
git config --global user.name EMCCMD-CI
git commit -m ":airplane: New final release v $integer_version" -m "[ci skip]"

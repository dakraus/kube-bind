#!/usr/bin/env bash

# Copyright 2024 The Kube Bind Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)

CRD_REF_DOCS_VERSION="${CRD_REF_DOCS_VERSION:-v0.1.0}"

SOURCE="${REPO_ROOT}/sdk/apis"
DESTINATION="${REPO_ROOT}/docs/content/reference/api"
mkdir -p "${DESTINATION}"

# Generate new content
go run github.com/elastic/crd-ref-docs@${CRD_REF_DOCS_VERSION} \
  --source-path "${SOURCE}" \
  --max-depth 10 \
  --renderer markdown \
  --templates-dir "${REPO_ROOT}/docs/generators/api-ref/templates" \
  --config "${REPO_ROOT}/docs/generators/api-ref/config.yaml" \
  --output-mode group \
  --output-path "${DESTINATION}"

# Organize APIs by group
for file in ${DESTINATION}/*.md; do
  filename=$(basename $file)
  apigroup=$(basename $filename .md)
  mkdir -p "${DESTINATION}/${apigroup}"

  csplit "${file}" \
    --elide-empty-files \
    --suppress-matched \
    --prefix="${DESTINATION}/${apigroup}/zz_generated." \
    --suffix-format='%03d.md' \
    '/<!-- SPLIT -->/' '{*}'

  for generated_file in ${DESTINATION}/${apigroup}/zz_generated.*.md; do
    kind=$(sed --quiet 's/^title: \(.*\)$/\1/p' $generated_file)
    mv "${generated_file}" "${DESTINATION}/${apigroup}/${kind@L}.md"
  done

  rm "${file}"
done

# Generate a .pages config file to override title case being applied to
# folder names by default (https://github.com/mkdocs/mkdocs/issues/2086)
echo "nav:" > ${DESTINATION}/.pages
for dir in ${DESTINATION}/*/; do
    if [ -d "${dir}" ]; then
        echo ${dir}
    fi
    apigroup=$(basename $dir)
    echo "  - ${apigroup}: ${apigroup}" >> ${DESTINATION}/.pages
done

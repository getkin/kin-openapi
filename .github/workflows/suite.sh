#!/bin/bash -eu

set -o pipefail

OAI_COMMIT=${OAI_COMMIT:-c3ac262c8e4b41bdc9da187dd6c7846981951ab6} # On 2024-06-10
OAI_SLUG=OAI-OpenAPI-Specification
OAI_DIR=$OAI_SLUG-${OAI_COMMIT:0:7}/examples

if ! [[ -d $OAI_DIR ]]; then
	curl -fSLo $OAI_SLUG.tar.gz https://github.com/OAI/OpenAPI-Specification/tarball/"$OAI_COMMIT"
	tar zxf $OAI_SLUG.tar.gz
	rm $OAI_SLUG.tar.gz
fi

APISGURU_COMMIT=${APISGURU_COMMIT:-226773819e337e6c413d1be91b26b33111211768} # On 2022-10-17
APISGURU_SLUG=APIs-guru-openapi-directory
APISGURU_DIR=$APISGURU_SLUG-${APISGURU_COMMIT:0:7}/APIs

if ! [[ -d $APISGURU_DIR ]]; then
	curl -fSLo $APISGURU_SLUG.tar.gz https://github.com/APIs-guru/openapi-directory/tarball/"$APISGURU_COMMIT"
	tar zxf $APISGURU_SLUG.tar.gz
	rm $APISGURU_SLUG.tar.gz
fi
# Ignore documents that are too large / take too long
rm -f $APISGURU_DIR/microsoft.com/graph/beta/openapi.yaml # >30MB
rm -f $APISGURU_DIR/microsoft.com/graph/v1.0/openapi.yaml # >15MB

cat <<EOF
name: suite

on:
  push:
  pull_request:
  schedule:
  - cron: '00 1 * * 1'  # At 01:00 on Mondays.
  workflow_dispatch:

jobs:
  meta-check:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - run: .github/workflows/suite.sh | tee .github/workflows/suite.yml
    - run: git --no-pager diff --exit-code

EOF


jobid=0
job() {
	local gen=${1}; shift
	local args=${1:-}
	((jobid++)) || true
	suite=suite_$(printf '%02d' $jobid)

cat <<EOF
  $suite:
    env:
      GITHUB_TOKEN: \${{ secrets.GITHUB_TOKEN }}
      GO111MODULE: 'on'
      CGO_ENABLED: '0'
    strategy:
      matrix:
        go: ['1.x']
        os:
        - ubuntu-latest
       #- windows-latest
       #- macos-latest
    runs-on: \${{ matrix.os }}
    defaults:
      run:
        shell: bash
    name: $suite on \${{ matrix.os }}
    steps:

    - uses: actions/setup-go@v2
      with:
        go-version: \${{ matrix.go }}

    - id: go-cache-paths
      run: |
        echo "::set-output name=go-build::\$(go env GOCACHE)"
        echo "::set-output name=go-mod::\$(go env GOMODCACHE)"
    - run: echo \${{ steps.go-cache-paths.outputs.go-build }}
    - run: echo \${{ steps.go-cache-paths.outputs.go-mod }}

    - name: Go Build Cache
      uses: actions/cache@v3
      with:
        path: \${{ steps.go-cache-paths.outputs.go-build }}
        key: \${{ runner.os }}-go-\${{ matrix.go }}-build-\${{ hashFiles('**/go.sum') }}

    - name: Go Mod Cache (go>=1.15)
      uses: actions/cache@v3
      with:
        path: \${{ steps.go-cache-paths.outputs.go-mod }}
        key: \${{ runner.os }}-go-\${{ matrix.go }}-mod-\${{ hashFiles('**/go.sum') }}

    - uses: actions/checkout@v2

    - name: Test suite fixtures cache
      uses: actions/cache@v3
      with:
        path: |
          APIs-guru-openapi-directory-*
          OAI-OpenAPI-Specification-*
        key: \${{ runner.os }}-go-\${{ matrix.go }}-suite-\${{ hashFiles('.github/workflows/suite.sh') }}

    - name: Maybe fetch fixtures
      run: |
        if ! [[ -d $OAI_DIR ]]; then
          curl -fSLo $OAI_SLUG.tar.gz https://github.com/OAI/OpenAPI-Specification/tarball/"$OAI_COMMIT"
          tar zxf $OAI_SLUG.tar.gz
          rm $OAI_SLUG.tar.gz
          ls -lha $OAI_DIR
        fi
        if ! [[ -d $APISGURU_DIR ]]; then
          curl -fSLo $APISGURU_SLUG.tar.gz https://github.com/APIs-guru/openapi-directory/tarball/"$APISGURU_COMMIT"
          tar zxf $APISGURU_SLUG.tar.gz
          rm $APISGURU_SLUG.tar.gz
          ls -lha $APISGURU_DIR
        fi

    - name: Ignore documents that are too large / take too long
      run: |
        rm -f $APISGURU_DIR/microsoft.com/graph/beta/openapi.yaml # >30MB
        rm -f $APISGURU_DIR/microsoft.com/graph/v1.0/openapi.yaml # >15MB

    - name: Build the tool
      run: go build -o . ./cmd/validate

    - name: "$gen"
      run: |
        set -x
        touch _
        while read -r file; do
          time ./validate -n 99 $args "\$file" || echo "\$file" >>_
        done < <($gen)

    - name: Compare with expected results
      run: |
        cat >__ <<EO_
EOF

while read -r file; do
  printf '        %s\n' "$file"
done

cat <<EOF
        EO_
        diff -u __ _

EOF

}

###

# TODO: drop these once documents are fixed upstream
# * https://github.com/APIs-guru/openapi-directory/pull/1064

###

job "ls $OAI_DIR/v2.0/json/*.json $OAI_DIR/v2.0/json/petstore-separate/spec/swagger.json" <<EOF
EOF

job "ls $OAI_DIR/v2.0/yaml/*.yaml $OAI_DIR/v2.0/yaml/petstore-separate/spec/swagger.yaml" <<EOF
EOF

job "find $OAI_DIR/v3.0 -type f | sort" <<EOF
EOF

job "find $OAI_DIR/v3.1 -type f | sort" <<EOF
$OAI_DIR/v3.1/non-oauth-scopes.json
$OAI_DIR/v3.1/non-oauth-scopes.yaml
$OAI_DIR/v3.1/webhook-example.json
$OAI_DIR/v3.1/webhook-example.yaml
EOF

###

job "find $APISGURU_DIR -type f -name swagger.yaml | sort | grep -F azure.com | grep -F network" <<EOF
EOF

job "find $APISGURU_DIR -type f -name swagger.yaml | sort | grep -F azure.com | grep -vF network" <<EOF
EOF

job "find $APISGURU_DIR -type f -name swagger.yaml | sort | grep -vF azure.com" <<EOF
$APISGURU_DIR/quarantine.country/1.0/swagger.yaml
EOF

# Sharded on subfolders:

list_shards() {
  cat .github/workflows/suite.sh | \grep -F '| grep -F' | \grep -vF swagger.yaml | sed -E 's%.+grep -F ([^"]+)".+%\1%g' | sort
}

shards=()
shards+=("adyen")
shards+=("amazonaws")
shards+=("apideck")
shards+=("apisetu")
shards+=("docusign")
shards+=("dracoon")
shards+=("ebay")
shards+=("github")
shards+=("google")
shards+=("here")
shards+=("interzoid")
shards+=("loket")
shards+=("microsoft")
shards+=("nexmo")
shards+=("pandascore")
shards+=("sportsdata")
shards+=("stripe")
shards+=("telnyx")
shards+=("twilio")
shards+=("zoom")
shards+=("zuora")
diff -u <(echo "${shards[@]}" | tr ' ' '\n') <(list_shards)

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F adyen" '--patterns=false' <<EOF
$APISGURU_DIR/adyen.com/AccountService/3/openapi.yaml
$APISGURU_DIR/adyen.com/AccountService/4/openapi.yaml
$APISGURU_DIR/adyen.com/AccountService/5/openapi.yaml
$APISGURU_DIR/adyen.com/AccountService/6/openapi.yaml
$APISGURU_DIR/adyen.com/BalancePlatformService/1/openapi.yaml
$APISGURU_DIR/adyen.com/BinLookupService/40/openapi.yaml
$APISGURU_DIR/adyen.com/BinLookupService/50/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/37/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/40/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/41/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/46/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/49/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/50/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/51/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/52/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/53/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/64/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/65/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/66/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/67/openapi.yaml
$APISGURU_DIR/adyen.com/CheckoutService/68/openapi.yaml
$APISGURU_DIR/adyen.com/FundService/3/openapi.yaml
$APISGURU_DIR/adyen.com/MarketPayNotificationService/3/openapi.yaml
$APISGURU_DIR/adyen.com/MarketPayNotificationService/4/openapi.yaml
$APISGURU_DIR/adyen.com/MarketPayNotificationService/5/openapi.yaml
$APISGURU_DIR/adyen.com/MarketPayNotificationService/6/openapi.yaml
$APISGURU_DIR/adyen.com/NotificationConfigurationService/1/openapi.yaml
$APISGURU_DIR/adyen.com/NotificationConfigurationService/2/openapi.yaml
$APISGURU_DIR/adyen.com/NotificationConfigurationService/3/openapi.yaml
$APISGURU_DIR/adyen.com/NotificationConfigurationService/4/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/25/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/30/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/40/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/46/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/49/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/50/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/51/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/52/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/64/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/67/openapi.yaml
$APISGURU_DIR/adyen.com/PaymentService/68/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/30/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/40/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/46/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/49/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/50/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/51/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/52/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/64/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/67/openapi.yaml
$APISGURU_DIR/adyen.com/PayoutService/68/openapi.yaml
$APISGURU_DIR/adyen.com/RecurringService/25/openapi.yaml
$APISGURU_DIR/adyen.com/RecurringService/30/openapi.yaml
$APISGURU_DIR/adyen.com/RecurringService/40/openapi.yaml
$APISGURU_DIR/adyen.com/RecurringService/49/openapi.yaml
$APISGURU_DIR/adyen.com/RecurringService/67/openapi.yaml
$APISGURU_DIR/adyen.com/RecurringService/68/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F amazonaws" '--patterns=false' <<EOF
$APISGURU_DIR/amazonaws.com/application-autoscaling/2016-02-06/openapi.yaml
$APISGURU_DIR/amazonaws.com/devicefarm/2015-06-23/openapi.yaml
$APISGURU_DIR/amazonaws.com/dynamodb/2012-08-10/openapi.yaml
$APISGURU_DIR/amazonaws.com/ec2/2016-11-15/openapi.yaml
$APISGURU_DIR/amazonaws.com/ecr/2015-09-21/openapi.yaml
$APISGURU_DIR/amazonaws.com/eks/2017-11-01/openapi.yaml
$APISGURU_DIR/amazonaws.com/elasticfilesystem/2015-02-01/openapi.yaml
$APISGURU_DIR/amazonaws.com/fsx/2018-03-01/openapi.yaml
$APISGURU_DIR/amazonaws.com/iam/2010-05-08/openapi.yaml
$APISGURU_DIR/amazonaws.com/inspector/2016-02-16/openapi.yaml
$APISGURU_DIR/amazonaws.com/lex-models/2017-04-19/openapi.yaml
$APISGURU_DIR/amazonaws.com/organizations/2016-11-28/openapi.yaml
$APISGURU_DIR/amazonaws.com/polly/2016-06-10/openapi.yaml
$APISGURU_DIR/amazonaws.com/s3/2006-03-01/openapi.yaml
$APISGURU_DIR/amazonaws.com/secretsmanager/2017-10-17/openapi.yaml
$APISGURU_DIR/amazonaws.com/servicediscovery/2017-03-14/openapi.yaml
$APISGURU_DIR/amazonaws.com/snowball/2016-06-30/openapi.yaml
$APISGURU_DIR/amazonaws.com/ssm-contacts/2021-05-03/openapi.yaml
$APISGURU_DIR/amazonaws.com/storagegateway/2013-06-30/openapi.yaml
$APISGURU_DIR/amazonaws.com/streams.dynamodb/2012-08-10/openapi.yaml
$APISGURU_DIR/amazonaws.com/waf-regional/2016-11-28/openapi.yaml
$APISGURU_DIR/amazonaws.com/waf/2015-08-24/openapi.yaml
$APISGURU_DIR/amazonaws.com/wafv2/2019-07-29/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F apideck" '--patterns=false' <<EOF
$APISGURU_DIR/apideck.com/accounting/8.70.2/openapi.yaml
$APISGURU_DIR/apideck.com/hris/8.70.2/openapi.yaml
$APISGURU_DIR/apideck.com/pos/8.70.2/openapi.yaml
$APISGURU_DIR/apideck.com/sms/8.70.2/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F apisetu" '--patterns=false' <<EOF
$APISGURU_DIR/apisetu.gov.in/aaharjh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/acko/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/agtripura/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/aharakar/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/aiimsmangalagiri/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/aiimspatna/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/aiimsrishikesh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/aktu/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/apmcservices/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/asrb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bajajallianz/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bajajallianzlife/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/barti/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bharatpetroleum/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bhartiaxagi/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bhavishya/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/biharboard/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bput/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/bsehr/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/cbse/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/cgbse/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/chennaicorp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/chitkarauniversity/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/cholainsurance/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/cisce/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/civilsupplieskerala/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/cpctmp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/csc/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/dbraitandaman/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/dgecerttn/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/dgft/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/dhsekerala/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ditarunachal/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ditch/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/dittripura/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/duexam/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistrictandaman/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistricthp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistrictkerala/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistrictodisha/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistrictodishasp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistrictpb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/edistrictup/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ehimapurtihp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/enibandhanjh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/epfindia/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/epramanhp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/eservicearunachal/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/fsdhr/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/futuregenerali/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/gadbih/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/gauhati/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/gbshse/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/geetanjaliuniv/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/gmch/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/goawrd/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/godigit/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/gujaratvidyapith/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hindustanpetroleum/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hpayushboard/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hpbose/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hppanchayat/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hpsbys/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hpsssb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hptechboard/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hsbte/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/hsscboardmh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/icicilombard/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/iciciprulife/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/icsi/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/igrmaharashtra/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/insvalsura/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/iocl/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/issuer/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/jac/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/jeecup/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/jharsewa/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/jnrmand/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/juit/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/keralapsc/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/kiadb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/kkhsou/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/kotakgeneralinsurance/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/kseebkr/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ktech/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/labourbih/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/landrecordskar/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/lawcollegeandaman/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/legalmetrologyup/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/licindia/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/maxlifeinsurance/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mbose/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mbse/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mcimindia/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/meark/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mizoramlesde/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mizorampolice/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mpmsu/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mppmc/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/mriu/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/msde/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/municipaladmin/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/nationalinsurance/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ncert/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/negd/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/neilit/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/newindia/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/niesbud/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/nios/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/nitap/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/nitp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/npsailu/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/nsdcindia/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/orientalinsurance/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pan/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pareekshabhavanker/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pblabour/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pgimer/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/phedharyana/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pmjay/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pramericalife/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/pseb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/puekar/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/punjabteched/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/rajasthandsa/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/rajasthanrajeduboard/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/reliancegeneral/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/revenueassam/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/revenueodisha/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/sainikwelfarepud/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/saralharyana/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/sbigeneral/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/scvtup/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/sebaonline/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/statisticsrajasthan/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/swavlambancard/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/tataaia/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/tataaig/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/tbse/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transport/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportan/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportap/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportar/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportas/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportbr/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportcg/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportdd/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportdh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportdl/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportga/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportgj/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transporthp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transporthr/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportjh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportjk/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportka/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportkl/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportld/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportmh/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportml/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportmn/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportmp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportmz/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportnl/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportod/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportpb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportpy/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportrj/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportsk/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transporttn/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transporttr/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportts/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportuk/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportup/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/transportwb/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ubseuk/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/ucobank/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/uiic/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/upmsp/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/vhseker/3.0.0/openapi.yaml
$APISGURU_DIR/apisetu.gov.in/vssut/3.0.0/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F docusign" '--patterns=false' <<EOF
$APISGURU_DIR/docusign.net/v2.1/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F dracoon" '--patterns=false' <<EOF
$APISGURU_DIR/dracoon.team/4.29.1/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F ebay" '--patterns=false' <<EOF
$APISGURU_DIR/ebay.com/sell-account/v1.6.3/openapi.yaml
$APISGURU_DIR/ebay.com/sell-compliance/1.4.1/openapi.yaml
$APISGURU_DIR/ebay.com/sell-feed/v1.3.1/openapi.yaml
$APISGURU_DIR/ebay.com/sell-fulfillment/v1.19.9/openapi.yaml
$APISGURU_DIR/ebay.com/sell-logistics/v1_beta.0.0/openapi.yaml
$APISGURU_DIR/ebay.com/sell-marketing/v1.10.0/openapi.yaml
$APISGURU_DIR/ebay.com/sell-negotiation/v1.1.0/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F github" '--patterns=false' <<EOF
$APISGURU_DIR/github.com/ghes-2.18/1.1.4/openapi.yaml
$APISGURU_DIR/github.com/ghes-2.19/1.1.4/openapi.yaml
$APISGURU_DIR/github.com/ghes-2.20/1.1.4/openapi.yaml
$APISGURU_DIR/github.com/ghes-2.21/1.1.4/openapi.yaml
$APISGURU_DIR/github.com/ghes-2.22/1.1.4/openapi.yaml
$APISGURU_DIR/github.com/ghes-3.0/1.1.4/openapi.yaml
$APISGURU_DIR/github.com/ghes-3.1/1.1.4/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F google" '--patterns=false' <<EOF
$APISGURU_DIR/googleapis.com/accessapproval/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/accesscontextmanager/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/admin/datatransfer_v1/openapi.yaml
$APISGURU_DIR/googleapis.com/admin/directory_v1/openapi.yaml
$APISGURU_DIR/googleapis.com/admin/reports_v1/openapi.yaml
$APISGURU_DIR/googleapis.com/admob/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/adsense/v1.4/openapi.yaml
$APISGURU_DIR/googleapis.com/adsense/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/appengine/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/appengine/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/artifactregistry/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/artifactregistry/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/bigqueryreservation/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/bigqueryreservation/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/bigtableadmin/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/billingbudgets/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/binaryauthorization/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/blogger/v3/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudasset/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudasset/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudasset/v1p1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudasset/v1p4beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudasset/v1p7beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudbilling/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudbilling/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudbuild/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudbuild/v1alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudbuild/v1alpha2/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudbuild/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudfunctions/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudfunctions/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudfunctions/v2alpha/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudfunctions/v2beta/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudidentity/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudresourcemanager/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudresourcemanager/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudresourcemanager/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudresourcemanager/v3/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudscheduler/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudshell/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudtasks/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudtasks/v2beta3/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudtrace/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/cloudtrace/v2beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/composer/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/container/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/container/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/containeranalysis/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/containeranalysis/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/contentwarehouse/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/datacatalog/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/datafusion/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/datastore/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/datastore/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/datastore/v1beta3/openapi.yaml
$APISGURU_DIR/googleapis.com/deploymentmanager/alpha/openapi.yaml
$APISGURU_DIR/googleapis.com/deploymentmanager/v2beta/openapi.yaml
$APISGURU_DIR/googleapis.com/dfareporting/v3.3/openapi.yaml
$APISGURU_DIR/googleapis.com/dfareporting/v3.4/openapi.yaml
$APISGURU_DIR/googleapis.com/dfareporting/v3.5/openapi.yaml
$APISGURU_DIR/googleapis.com/dfareporting/v4/openapi.yaml
$APISGURU_DIR/googleapis.com/discovery/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/displayvideo/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/displayvideo/v1dev/openapi.yaml
$APISGURU_DIR/googleapis.com/displayvideo/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/dns/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/dns/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/dns/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/doubleclickbidmanager/v1.1/openapi.yaml
$APISGURU_DIR/googleapis.com/doubleclickbidmanager/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/file/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/file/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/firebaseappcheck/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/firebaseappcheck/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/firebasehosting/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/firebasehosting/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/firebaseml/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/firestore/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/firestore/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/firestore/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/gameservices/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/gameservices/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/genomics/v2alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/healthcare/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/healthcare/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/iam/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/iam/v2beta/openapi.yaml
$APISGURU_DIR/googleapis.com/iap/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/ideahub/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/identitytoolkit/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/identitytoolkit/v3/openapi.yaml
$APISGURU_DIR/googleapis.com/jobs/v3/openapi.yaml
$APISGURU_DIR/googleapis.com/jobs/v3p1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/jobs/v4/openapi.yaml
$APISGURU_DIR/googleapis.com/language/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/language/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/managedidentities/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/managedidentities/v1alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/managedidentities/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/memcache/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/monitoring/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/monitoring/v3/openapi.yaml
$APISGURU_DIR/googleapis.com/networkmanagement/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/networksecurity/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/osconfig/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/osconfig/v1alpha/openapi.yaml
$APISGURU_DIR/googleapis.com/osconfig/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/oslogin/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/oslogin/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/policytroubleshooter/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/policytroubleshooter/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/privateca/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/pubsub/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/pubsub/v1beta1a/openapi.yaml
$APISGURU_DIR/googleapis.com/pubsub/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/recommender/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/redis/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/redis/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/run/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/run/v1alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/run/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/runtimeconfig/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/secretmanager/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/securitycenter/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/securitycenter/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/securitycenter/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/securitycenter/v1p1alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/serviceconsumermanagement/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/servicecontrol/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/servicecontrol/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/servicedirectory/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/servicenetworking/v1beta/openapi.yaml
$APISGURU_DIR/googleapis.com/serviceusage/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/speech/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/speech/v1p1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/speech/v2beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/tagmanager/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/texttospeech/v1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/tpu/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/tpu/v1alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/tpu/v2alpha1/openapi.yaml
$APISGURU_DIR/googleapis.com/translate/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/translate/v3beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/verifiedaccess/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/verifiedaccess/v2/openapi.yaml
$APISGURU_DIR/googleapis.com/videointelligence/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/videointelligence/v1beta2/openapi.yaml
$APISGURU_DIR/googleapis.com/videointelligence/v1p1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/videointelligence/v1p3beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/vision/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/vision/v1p1beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/vision/v1p2beta1/openapi.yaml
$APISGURU_DIR/googleapis.com/websecurityscanner/v1/openapi.yaml
$APISGURU_DIR/googleapis.com/websecurityscanner/v1beta/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F here" '--patterns=false' <<EOF
$APISGURU_DIR/amentum.space/atmosphere/1.1.1/openapi.yaml
$APISGURU_DIR/here.com/positioning/2.1.1/openapi.yaml
$APISGURU_DIR/here.com/tracking/2.1.58/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F interzoid" '--patterns=false' <<EOF
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F loket" '--patterns=false' <<EOF
$APISGURU_DIR/loket.nl/V2/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F microsoft" '--patterns=false' <<EOF
$APISGURU_DIR/microsoft.com/cognitiveservices-ComputerVision/2.0/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-ComputerVision/2.1/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Ocr/2.0/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Ocr/2.1/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Prediction/2.0/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Prediction/3.0/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Training/2.0/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Training/2.1/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Training/2.2/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Training/3.0/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Training/3.1/openapi.yaml
$APISGURU_DIR/microsoft.com/cognitiveservices-Training/3.2/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F nexmo" '--patterns=false' <<EOF
$APISGURU_DIR/nexmo.com/account/1.0.4/openapi.yaml
$APISGURU_DIR/nexmo.com/audit/1.0.4/openapi.yaml
$APISGURU_DIR/nexmo.com/conversation.v2/1.0.1/openapi.yaml
$APISGURU_DIR/nexmo.com/conversation/2.0.1/openapi.yaml
$APISGURU_DIR/nexmo.com/dispatch/0.3.4/openapi.yaml
$APISGURU_DIR/nexmo.com/messages-olympus/1.2.3/openapi.yaml
$APISGURU_DIR/nexmo.com/number-insight/1.2.1/openapi.yaml
$APISGURU_DIR/nexmo.com/reports/2.2.2/openapi.yaml
$APISGURU_DIR/nexmo.com/subaccounts/1.0.8/openapi.yaml
$APISGURU_DIR/nexmo.com/verify/1.2.3/openapi.yaml
$APISGURU_DIR/nexmo.com/voice/1.3.8/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F pandascore" '--patterns=false' <<EOF
$APISGURU_DIR/pandascore.co/2.23.1/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F sportsdata" '--patterns=false' <<EOF
$APISGURU_DIR/sportsdata.io/mlb-v3-rotoballer-articles/1.0/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F stripe" '--patterns=false' <<EOF
$APISGURU_DIR/stripe.com/2020-08-27/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F telnyx" '--patterns=false' <<EOF
$APISGURU_DIR/telnyx.com/2.0.0/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F twilio" '--patterns=false' <<EOF
$APISGURU_DIR/twilio.com/api/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_chat_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_chat_v2/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_chat_v3/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_conversations_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_insights_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_ip_messaging_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_ip_messaging_v2/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_media_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_messaging_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_numbers_v2/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_preview/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_supersim_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_sync_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_taskrouter_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_trusthub_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_verify_v2/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_video_v1/1.36.0/openapi.yaml
$APISGURU_DIR/twilio.com/twilio_wireless_v1/1.36.0/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F zoom" '--patterns=false' <<EOF
$APISGURU_DIR/zoom.us/2.0.0/openapi.yaml
EOF

job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -F zuora" '--patterns=false' <<EOF
$APISGURU_DIR/zuora.com/2021-08-20/openapi.yaml
EOF

rest=''
for shard in "${shards[@]}"; do
  rest="$rest|$shard"
done
job "find $APISGURU_DIR -type f -name openapi.yaml | sort | grep -vE '$rest'" '--patterns=false' <<EOF
EOF

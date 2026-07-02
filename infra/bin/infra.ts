import 'dotenv/config'

import * as cdk from 'aws-cdk-lib'

import { RecipeStack } from '../lib/recipe-stack'

const app = new cdk.App()

new RecipeStack(app, 'RecipeStack', {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    // Neon(無料枠)が東京に対応していないため、DB と同居できるシンガポールに置く
    // (adr/infra/0002 #2)。CloudFront が前段にあるため体感差は小さい。
    region: 'ap-southeast-1',
  },
})

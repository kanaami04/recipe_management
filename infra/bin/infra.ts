import 'dotenv/config'

import * as cdk from 'aws-cdk-lib'

import { RecipeStack } from '../lib/recipe-stack'

const app = new cdk.App()

new RecipeStack(app, 'RecipeStack', {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    // DB(Supabase)と同じ東京リージョンに置く。
    region: 'ap-northeast-1',
  },
})

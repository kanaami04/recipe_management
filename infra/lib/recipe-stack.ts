import { CfnOutput, Duration, RemovalPolicy, Stack, type StackProps } from 'aws-cdk-lib'
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront'
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins'
import * as lambda from 'aws-cdk-lib/aws-lambda'
import * as logs from 'aws-cdk-lib/aws-logs'
import * as s3 from 'aws-cdk-lib/aws-s3'
import * as s3deploy from 'aws-cdk-lib/aws-s3-deployment'
import type { Construct } from 'constructs'

// S3 + CloudFront + Lambda(Function URL) の最小最安構成 (adr/infra/0001)。
// フロントと API を CloudFront で同一オリジンに束ね、既存の Cookie/CSRF 設計
// (adr/api/0004, adr/frontend/0004) を無改修で成立させる。
export class RecipeStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props)

    // ---- Lambda (Go + Lambda Web Adapter) --------------------------------
    // Go バイナリは `bootstrap` の名前で infra/dist/api/ に事前ビルドする
    // (mise run build-lambda)。LWA レイヤーが Lambda イベントと HTTP を変換する
    // ため、アプリは通常の常駐 HTTP サーバのまま動く (adr/infra/0001 #2)。
    const api = new lambda.Function(this, 'Api', {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      architecture: lambda.Architecture.ARM_64,
      handler: 'bootstrap',
      code: lambda.Code.fromAsset('dist/api'),
      memorySize: 256,
      timeout: Duration.seconds(30),
      layers: [
        lambda.LayerVersion.fromLayerVersionArn(
          this,
          'LambdaWebAdapter',
          `arn:aws:lambda:${this.region}:753240598075:layer:LambdaAdapterLayerArm64:28`,
        ),
      ],
      environment: {
        PORT: '8080',
        // コールドスタート毎の DDL を避ける。マイグレーションはローカルから
        // session pooler の DSN で `go run . -migrate` (adr/infra/0002 #3)。
        AUTO_MIGRATE: 'false',
        COOKIE_SECURE: 'true',
        LOG_LEVEL: 'info',
        LOG_FORMAT: 'json',
        DATABASE_URL: requireEnv('DATABASE_URL'),
        JWT_SECRET: requireEnv('JWT_SECRET'),
        ORIGIN_VERIFY_SECRET: requireEnv('ORIGIN_VERIFY_SECRET'),
        // 本番は同一オリジンのため CORS は実質発火しない。初回デプロイ後に
        // 判明する CloudFront ドメインを -c corsOrigin=... で一度だけ反映する。
        CORS_ORIGIN: this.node.tryGetContext('corsOrigin') ?? 'https://placeholder.invalid',
      },
      logGroup: new logs.LogGroup(this, 'ApiLogs', {
        retention: logs.RetentionDays.ONE_MONTH,
        removalPolicy: RemovalPolicy.DESTROY,
      }),
    })

    // Function URL は認証なしで公開し、CloudFront 経由以外はアプリ層の
    // X-Origin-Verify 検証で遮断する (adr/infra/0001 #3)。
    const apiUrl = api.addFunctionUrl({ authType: lambda.FunctionUrlAuthType.NONE })

    // ---- S3 (SPA 静的アセット) -------------------------------------------
    const web = new s3.Bucket(this, 'Web', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      enforceSSL: true,
      removalPolicy: RemovalPolicy.DESTROY,
      autoDeleteObjects: true,
    })

    // ---- CloudFront ------------------------------------------------------
    const spaRewrite = new cloudfront.Function(this, 'SpaRewrite', {
      runtime: cloudfront.FunctionRuntime.JS_2_0,
      code: cloudfront.FunctionCode.fromFile({ filePath: 'functions/spa-rewrite.js' }),
      comment: 'SPA fallback: extensionless URI -> /index.html',
    })

    const distribution = new cloudfront.Distribution(this, 'Distribution', {
      comment: 'recipe-management',
      // 日本のエッジを含める(北米/欧州のみの PRICE_CLASS_100 だと遅くなる)。
      priceClass: cloudfront.PriceClass.PRICE_CLASS_200,
      defaultRootObject: 'index.html',
      defaultBehavior: {
        origin: origins.S3BucketOrigin.withOriginAccessControl(web),
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        compress: true,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
        functionAssociations: [
          { function: spaRewrite, eventType: cloudfront.FunctionEventType.VIEWER_REQUEST },
        ],
      },
      additionalBehaviors: {
        '/api/*': {
          origin: new origins.FunctionUrlOrigin(apiUrl, {
            // CloudFront 経由の証明。アプリ層のミドルウェアがこれを検証する。
            customHeaders: { 'X-Origin-Verify': requireEnv('ORIGIN_VERIFY_SECRET') },
          }),
          allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
          viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.HTTPS_ONLY,
          // API はキャッシュしない。Cookie/Authorization を含む全ヘッダを転送するが、
          // Host だけは除外必須(転送すると Function URL のルーティングが壊れる)。
          cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
          originRequestPolicy: cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
        },
      },
    })

    // ---- フロントのデプロイ ----------------------------------------------
    // ハッシュ付きアセットは immutable 長期キャッシュ。
    const assetsDeployment = new s3deploy.BucketDeployment(this, 'WebAssets', {
      sources: [s3deploy.Source.asset('../frontend/build/client/assets')],
      destinationBucket: web,
      destinationKeyPrefix: 'assets',
      cacheControl: [s3deploy.CacheControl.fromString('public,max-age=31536000,immutable')],
      // 旧アセットを消さない(公開中の index.html / sw.js が参照している可能性がある)。
      prune: false,
    })
    // index.html / sw.js / manifest 等は no-cache(毎回再検証)。
    // ここを長期キャッシュにすると SW の更新が永遠に届かなくなる (adr/infra/0001 #4)。
    const rootDeployment = new s3deploy.BucketDeployment(this, 'WebRoot', {
      sources: [s3deploy.Source.asset('../frontend/build/client', { exclude: ['assets/**'] })],
      destinationBucket: web,
      cacheControl: [s3deploy.CacheControl.fromString('no-cache')],
      prune: false,
      distribution,
      distributionPaths: ['/*'],
    })
    // 新しい index.html が参照するアセットを先にアップロードする。
    rootDeployment.node.addDependency(assetsDeployment)

    // ---- Outputs ---------------------------------------------------------
    new CfnOutput(this, 'Url', { value: `https://${distribution.distributionDomainName}` })
    new CfnOutput(this, 'ApiFunctionUrl', {
      value: apiUrl.url,
      description: '直叩き確認用(X-Origin-Verify なしでは 403 になること)',
    })
  }
}

// 必須の環境変数を取得する。未設定なら synth 時点で失敗させる(infra/.env 参照)。
function requireEnv(key: string): string {
  const value = process.env[key]
  if (!value) {
    throw new Error(`環境変数 ${key} が未設定です。infra/.env.example を参照してください。`)
  }
  return value
}

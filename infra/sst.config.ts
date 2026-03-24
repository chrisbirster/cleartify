/// <reference path="./.sst/platform/config.d.ts" />

export default $config({
  app(input) {
    return {
      name: "cleartify-infra",
      home: "aws",
      removal: input?.stage === "production" ? "retain" : "remove",
      providers: {
        aws: {
          region: (process.env.AWS_REGION ?? "us-east-1") as any,
        },
      },
    };
  },
  async run() {
    const authHeaderName = (process.env.cleartify_EMAIL_API_AUTH_HEADER || "Authorization").trim();
    const sender = (process.env.cleartify_EMAIL_SENDER || "no-reply@cleartify.com").trim();
    const from = (process.env.cleartify_EMAIL_FROM || sender).trim();

    const emailApiKey = new sst.Secret("EmailApiKey");

    const email = new sst.aws.Email("cleartifyEmail", {
      sender,
    });

    const emailApi = new sst.aws.Function("EmailApi", {
      handler: "functions/email.handler",
      runtime: "nodejs20.x",
      timeout: "15 seconds",
      memory: "256 MB",
      url: true,
      link: [email, emailApiKey],
      environment: {
        EMAIL_SEND_AUTH_HEADER: authHeaderName,
        EMAIL_FROM: from,
      },
    });

    return {
      region: process.env.AWS_REGION ?? "us-east-1",
      sender,
      from,
      authHeaderName,
      emailApiBaseUrl: emailApi.url,
    };
  },
});

# ADR-006: Azure + .NET Aspire Architecture

## Status
Accepted

## Context
We need to decide on the technical architecture for the platform. The developer has strong experience with:
- Azure (Functions, Container Apps, etc.)
- .NET ecosystem
- .NET Aspire for local orchestration

## Decision

**Build on .NET Aspire with Azure as the deployment target.**

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Azure                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    .NET Aspire Application                       │    │
│  │                                                                   │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │    │
│  │  │   Web UI     │  │   REST API   │  │  Background Worker   │   │    │
│  │  │   (Blazor)   │  │  (Minimal    │  │  (Sync, Analysis)    │   │    │
│  │  │              │  │   APIs)      │  │                      │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘   │    │
│  │                                                                   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                              │                                           │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    Azure Functions                               │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │    │
│  │  │  OAuth       │  │  Timer       │  │  FIT Parser          │   │    │
│  │  │  Callbacks   │  │  Triggers    │  │  (Python isolated)   │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                              │                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌───────────┐   │
│  │ Blob Storage │  │  Cosmos DB   │  │ Service Bus  │  │  Azure    │   │
│  │ (FIT files)  │  │  (metadata)  │  │  (queues)    │  │  OpenAI   │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └───────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Breakdown

| Component | Technology | Responsibility |
|-----------|------------|----------------|
| **Web UI** | Blazor Server | Dashboard, activity viewer, settings |
| **API** | ASP.NET Minimal APIs | REST endpoints for all operations |
| **Worker** | .NET Worker Service | Background sync, LLM analysis |
| **OAuth Handler** | Azure Functions (.NET) | Stable callback URLs for providers |
| **FIT Parser** | Azure Functions (Python) | Leverage Python `fitdecode` library |
| **File Storage** | Azure Blob Storage | Raw FIT files with provenance metadata |
| **Database** | Cosmos DB | Activity metadata, user settings, sync state |
| **Message Queue** | Azure Service Bus | Decouple sync triggers from processing |
| **LLM** | Azure OpenAI + Semantic Kernel | Workout summaries, chat, suggestions |
| **Secrets** | Azure Key Vault | OAuth tokens, API keys |
| **Auth** | Azure Entra ID (B2C) | User authentication |

### Why This Stack?

| Decision | Rationale |
|----------|-----------|
| .NET Aspire | Excellent local dev experience, built-in observability, Aspire 9 is mature |
| Azure Functions for OAuth | Stable URLs without custom domain setup |
| Python for FIT parsing | `fitdecode` is the best FIT library; isolated function avoids polyglot complexity |
| Cosmos DB | Document model fits activity data; global distribution if needed later |
| Service Bus | Reliable queuing for async sync jobs; dead letter for failures |
| Semantic Kernel | .NET-native LLM SDK; works with Azure OpenAI |

### Local Development

Aspire orchestrates everything locally:

```csharp
// AppHost/Program.cs
var builder = DistributedApplication.CreateBuilder(args);

var cosmos = builder.AddAzureCosmosDB("cosmos")
    .RunAsEmulator();

var storage = builder.AddAzureStorage("storage")
    .RunAsEmulator()
    .AddBlobs("fit-files");

var serviceBus = builder.AddAzureServiceBus("messaging")
    .RunAsEmulator();

var api = builder.AddProject<Projects.FitSync_Api>("api")
    .WithReference(cosmos)
    .WithReference(storage);

var worker = builder.AddProject<Projects.FitSync_Worker>("worker")
    .WithReference(cosmos)
    .WithReference(storage)
    .WithReference(serviceBus);

var web = builder.AddProject<Projects.FitSync_Web>("web")
    .WithReference(api);

builder.Build().Run();
```

### Deployment

| Environment | Target | Notes |
|-------------|--------|-------|
| Local | Aspire with emulators | `dotnet run` in AppHost |
| Dev/Test | Azure Container Apps | Low-cost, scale to zero |
| Production | Azure Container Apps | Auto-scaling, managed |
| Functions | Azure Functions (Consumption) | Pay per execution |

### Cost Estimate (Low Usage)

| Service | Tier | Estimated Monthly |
|---------|------|-------------------|
| Container Apps | Consumption | $5-20 |
| Functions | Consumption | $0-5 |
| Cosmos DB | Serverless | $5-25 |
| Blob Storage | Hot | $1-5 |
| Service Bus | Basic | $0.05 |
| Azure OpenAI | Pay-per-token | $5-20 |
| **Total** | | **~$20-75/month** |

## Consequences

### Positive
- Mature, well-supported stack
- Excellent local development experience
- Azure-native security (Managed Identity, Key Vault)
- Scalable from hobby to production
- Strong observability out of the box

### Negative
- Azure lock-in (mitigated by clean architecture)
- Python function adds deployment complexity
- Learning curve for Aspire if unfamiliar

### Neutral
- Requires Azure subscription
- Cost scales with usage

## Related Decisions
- ADR-002: LLM Provider Strategy (Azure OpenAI as primary)
- ADR-003: FIT Files as Canonical Format

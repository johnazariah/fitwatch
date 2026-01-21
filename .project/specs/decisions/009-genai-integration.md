# ADR-009: GenAI Integration Points

## Status
Accepted

## Context
We want to provide LLM-powered analysis to help athletes improve. We need to decide:
- What AI features to build
- When/how they're triggered
- What data is sent to the LLM
- Privacy considerations

## Decision

**Integrate AI at natural touchpoints in the user journey. Start with automatic summaries, add interactive chat in Phase 2.**

### AI Features Roadmap

| Feature | Phase | Tier | Trigger | User Value |
|---------|-------|------|---------|------------|
| **Workout Summary** | MVP | Premium | Auto after sync | "What did I just do?" |
| **Weekly Digest** | MVP | Premium | Scheduled (Monday) | "How was my week?" |
| **Anomaly Detection** | MVP | Premium | Auto after sync | "Something seems off" |
| **Training Load** | P1 | Premium | Dashboard view | "Am I overtraining?" |
| **Cross-Platform Calibration** | P1 | Premium | On demand | "Is my Zwift power lying?" |
| **FTP Estimation** | P1 | Premium | On demand | "What's my FTP without testing?" |
| **Trend Insights** | P1 | Premium | Monthly | "How am I progressing?" |
| **Chat Interface** | P2 | Premium | User-initiated | "Ask anything" |
| **Workout Suggestions** | P2 | Premium | User-initiated | "What should I do next?" |
| **Plan Comparison** | P3 | Premium | If plan connected | "Am I following my plan?" |

> **Note:** All AI features require Premium tier ($10/year). Free tier provides sync-only functionality.

### Implementation: Semantic Kernel

Use Microsoft Semantic Kernel for LLM orchestration:

```csharp
// Setup
var kernel = Kernel.CreateBuilder()
    .AddAzureOpenAIChatCompletion(
        deploymentName: "gpt-4o",
        endpoint: config["AzureOpenAI:Endpoint"],
        apiKey: config["AzureOpenAI:ApiKey"])
    .Build();

// Add plugins for fitness-specific functions
kernel.Plugins.AddFromType<FitnessMetricsPlugin>();
kernel.Plugins.AddFromType<WorkoutHistoryPlugin>();
```

### Feature Details

#### 1. Workout Summary (MVP)

**Trigger:** Automatic after successful FIT file import

**Input to LLM:**
```json
{
  "activity_type": "cycling",
  "duration_minutes": 62,
  "distance_km": 35.2,
  "avg_power": 185,
  "normalized_power": 198,
  "avg_hr": 142,
  "max_hr": 168,
  "tss": 65,
  "intensity_factor": 0.78,
  "intervals": [
    {"duration": 300, "avg_power": 250, "type": "threshold"}
  ],
  "comparison": {
    "last_week_avg_power": 180,
    "last_month_avg_tss": 55
  }
}
```

**Prompt:**
```
You are a cycling coach analyzing a workout. Provide a 2-3 sentence summary 
that highlights:
- What type of workout this was (endurance, intervals, race, recovery)
- Key achievements or notable metrics
- One brief observation or suggestion

Keep it conversational and encouraging. Use cycling terminology appropriately.

Workout data:
{{$workoutData}}
```

**Output:**
> "Solid endurance ride with some punchy efforts. Your normalized power of 198W
> shows you were working harder than the average suggests — probably some hills
> or surges. Good consistency with your recent training load."

**Storage:** Summary stored with activity, regeneratable on demand.

#### 2. Weekly Digest (MVP)

**Trigger:** Scheduled job every Monday 8am (user timezone)

**Input:** All activities from past 7 days + trailing 4-week context

**Prompt:**
```
You are a cycling coach reviewing an athlete's week. Analyze their training 
and provide:

1. A 2-sentence summary of the week
2. Total volume vs previous weeks (improving, maintaining, or reducing)
3. One thing they did well
4. One suggestion for next week
5. Recovery status (fresh, tired, or at risk of overtraining)

Be specific to their actual data. Use cycling terminology.

This week's activities:
{{$thisWeek}}

Previous weeks for context:
{{$previousWeeks}}
```

**Delivery:** Email and/or dashboard notification

#### 3. Anomaly Detection (MVP)

**Trigger:** Auto after sync, before summary generation

**Check for:**
- Heart rate unusually high for power output (cardiac drift, illness, heat)
- Power unusually low for heart rate (fatigue, bonk)
- Missing data streams (sensor issues)
- Unusual workout pattern (3am ride?)

**Implementation:**
```csharp
public class AnomalyDetector
{
    public async Task<List<Anomaly>> DetectAnomalies(Activity activity, 
        UserBaseline baseline)
    {
        var anomalies = new List<Anomaly>();
        
        // Check efficiency factor (power:HR ratio)
        var ef = activity.Summary.NormalizedPower / activity.Summary.AvgHeartRate;
        if (ef < baseline.EfficiencyFactor * 0.85)
        {
            anomalies.Add(new Anomaly
            {
                Type = AnomalyType.LowEfficiency,
                Message = "Your power-to-HR ratio was lower than usual",
                Severity = Severity.Info
            });
        }
        
        // More checks...
        
        return anomalies;
    }
}
```

**If anomalies found:** Include in LLM prompt for contextual explanation.

#### 4. Chat Interface (Phase 2)

**Trigger:** User clicks "Ask about your training"

**Architecture:**
```
User Question
     │
     ▼
┌─────────────────────────────────────────┐
│           Semantic Kernel               │
│  ┌─────────────────────────────────┐   │
│  │     Function Calling            │   │
│  │  - GetRecentActivities()        │   │
│  │  - GetActivityDetails(id)       │   │
│  │  - CalculateTrends(metric)      │   │
│  │  - ComparePeriods(p1, p2)       │   │
│  └─────────────────────────────────┘   │
│                 │                       │
│                 ▼                       │
│  ┌─────────────────────────────────┐   │
│  │         Azure OpenAI            │   │
│  │    GPT-4o with function calling │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
     │
     ▼
Natural Language Response
```

**Example conversation:**
```
User: "How did my FTP progress this year?"

System: [Calls CalculateTrends("ftp", "year")]
        [Returns: Jan: 220W, Apr: 235W, Jul: 248W, Now: 255W]

AI: "Your FTP has improved from 220W in January to 255W now — 
     that's a 16% gain over the year! You saw the biggest jump 
     between April and July. Nice work staying consistent."
```

### Privacy Considerations

| Concern | Mitigation |
|---------|------------|
| GPS data sent to LLM | Strip GPS; only send metrics |
| Workout history size | Summarize historical data, don't send raw |
| Personal information | No names, locations, or identifiable data |
| Data retention | Azure OpenAI doesn't retain data (verify) |
| User control | Option to disable AI features |

**What we send to LLM:**
```json
{
  "workout_metrics": { ... },      // ✅ Safe
  "lap_summaries": [ ... ],        // ✅ Safe
  "gps_coordinates": [ ... ],      // ❌ NEVER
  "user_name": "...",              // ❌ NEVER  
  "device_serial": "..."           // ❌ NEVER
}
```

### Cost Estimation

| Feature | Calls/Week | Tokens/Call | Weekly Cost |
|---------|------------|-------------|-------------|
| Workout Summary | 5 | 500 | $0.02 |
| Weekly Digest | 1 | 1500 | $0.01 |
| Chat (if used) | 10 | 800 | $0.08 |
| **Total** | | | **~$0.50/user/week** |

At $2/user/month for AI, well within hobby budget.

## Premium Tier: AI Coaching

AI features are the **premium differentiator** in the business model:

| Tier | Features | Price |
|------|----------|-------|
| **Free** | Sync to Intervals.icu (all platforms) | $0 |
| **Premium** | AI coaching insights | $10/year |

### Business Model Flywheel

```
Free sync solves pain → User captures tokens → Activities accumulate
                                                      │
                                                      ▼
                        AI improves with more data ← Premium unlocks AI
                                    │
                                    ▼
                        User sees trends, gets hooked → Renews
```

The free tier captures users and data. The premium tier monetizes the accumulated context—AI insights get **more valuable** the longer a user is on the platform.

### Premium AI Features

#### 1. Post-Workout Analysis (Enhanced)

Beyond basic summaries, premium users get:

- **Cardiac drift detection**: Did HR creep up while power stayed flat? Sign of dehydration, heat, or low glycogen
- **Power:HR ratio (Efficiency Factor)**: Compare to baseline—are you fresh or fatigued?
- **Pacing analysis**: Did you blow up? Negative split or faded?
- **RPE vs actual output**: If we have RPE data, compare subjective feel to objective load

#### 2. Cross-Platform Calibration

Users riding both Zwift and MyWhoosh often see power discrepancies:

```
"Your average power on Zwift is consistently 8% higher than MyWhoosh 
for similar HR zones. This suggests either:
- Different trainer calibration between platforms
- Zwift's resistance model being more generous

Consider: Calibrate your trainer before each session, or treat 
platform-specific FTP values separately."
```

**Implementation:** Compare similar-effort rides (by HR zone distribution) across platforms. Flag systematic bias.

#### 3. Training Load Balance

Analyze polarized vs sweetspot distribution:

```
"Over the past 4 weeks, 65% of your time was in Zone 3 (tempo). 
Research suggests either going easier (Zone 2) or harder (Zone 4+) 
produces better adaptations than middle-zone work.

Try: Replace one tempo ride with 4x8min threshold intervals, 
and one with a 90-minute easy spin."
```

#### 4. FTP Estimation

From accumulated data, estimate FTP without formal testing:

- Look for 20-60 minute sustained efforts
- Analyze power curve from all activities
- Factor in recent training load (are you fresh or fatigued?)

```
"Based on your best 20-minute efforts and recent training load, 
your current FTP is likely 248-255W. Your last test was 8 weeks ago 
at 235W—time for a retest?"
```

#### 5. Trend Insights

Monthly and quarterly trend analysis:

- Volume trends (hours, TSS, distance)
- Intensity distribution shifts
- Power curve changes (5s, 1m, 5m, 20m, 60m)
- Weight-adjusted metrics if weight tracked

```
"Q4 Summary: 
- Volume up 12% vs Q3
- Threshold power improved 3.2%
- But your 5-second power dropped 8%—consider adding some sprints
- You're averaging 6.2 hours/week, up from 5.5 in Q3"
```

#### 6. Workout Suggestions

Based on training gaps and goals:

```
"Looking at your last 2 weeks:
- Lots of endurance and tempo ✓
- No VO2max work (0 minutes >106% FTP)
- Last threshold session was 10 days ago

Suggested this week:
1. Tuesday: 5x3min @ 115% FTP, 3min recovery
2. Thursday: 2x15min @ 95% FTP
3. Weekend: Long endurance ride"
```

### Cost Analysis at Scale

Using GPT-4o pricing (~$5/1M input tokens, $15/1M output tokens):

| Scenario | Token Usage | Cost/Analysis |
|----------|-------------|---------------|
| Workout summary | ~800 input, 200 output | $0.007 |
| Weekly digest | ~2000 input, 400 output | $0.016 |
| Trend analysis | ~3000 input, 500 output | $0.023 |

**At scale (100 users, 20 activities/month each):**
- 2000 workout summaries: $14
- 400 weekly digests: $6.40
- 100 monthly trends: $2.30
- **Total: ~$23/month**

With $10/year premium ($83/month revenue at 100 users), AI costs are ~28% of revenue—sustainable.

### Token Accumulation Advantage

The more activities a user syncs, the better their AI coaching:
- Fresh users: Generic advice based on single workout
- 3-month users: Can spot trends, compare periods
- 1-year users: Seasonal analysis, year-over-year comparisons

This creates a **switching cost**—leaving FitBridge means losing accumulated context.

## Consequences

### Positive
- AI adds clear user value
- Semantic Kernel handles complexity
- Privacy preserved by design
- Cost is minimal
- **Premium tier has clear value prop vs free**
- **Data accumulation creates moat and retention**

### Negative
- Requires Azure OpenAI subscription
- LLM quality varies; need good prompts
- Users may expect more than AI can deliver
- **Prompt engineering needed for each platform's data quirks**

### Alternative: Local LLM
If user runs Ollama locally, we can use that instead:
```csharp
kernel.AddOpenAIChatCompletion(
    modelId: "llama3.2",
    endpoint: new Uri("http://localhost:11434"),
    apiKey: "not-needed");
```

Lower quality but fully private.

## Related Decisions
- ADR-002: LLM Provider Strategy
- ADR-006: Azure + .NET Aspire Architecture
- ADR-019: Serverless Multi-Tenant Architecture (hosting context)

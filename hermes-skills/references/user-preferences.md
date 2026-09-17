# User Communication Preferences

## Directness
- No filler phrases ("Great question", "I'd be happy to", "Let me explain")
- No restating the request back
- No narrating tool calls the user can see
- Match length to weight of ask — one-line question gets one-line answer

## Blockers
- Report blockers clearly and STOP — do not retry endlessly
- When blocked, say exactly what's needed from the user
- Frustration is a first-class skill signal — embed the lesson

## Architecture Corrections
- User understands K8s architecture deeply — don't over-explain basics
- When corrected, fix it immediately and move on
- Don't drift from the stated goal

## Ingress Architecture (corrected by user)
- Single ingress-nginx controller watches ALL namespaces
- Per-namespace Ingress resources define routing rules for that namespace
- NOT one Ingress to rule them all
- Staging Ingress strips `/staging` prefix via rewrite-target annotation
- Production Ingress has no rewrite
- Both use the same controller (single api gateway)

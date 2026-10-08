## Parent

#1

## What to build

A normalization and composite scoring engine in Go that processes all snapshots within a crawl round batch, calculates Expected Return ($\text{Price} \times \text{Commission Rate}\%$) per snapshot, applies batch Min-Max Normalization (0–100) with 99th percentile clipping across Sales Velocity, Commission Rate, and Expected Return, and computes the composite Winning Score using 50/30/20 weights.

## Acceptance criteria

- [x] Expected Return in THB is calculated accurately using snapshot-specific price and commission rate.
- [x] Outlier clipping at the 99th percentile prevents extreme viral spikes or exorbitant prices from compressing normal scores.
- [x] Min-Max normalization scales Velocity, Commission Rate, and Expected Return strictly into [0, 100].
- [x] Composite Winning Score is computed: $(0.50 \times \text{NormVelocity}) + (0.30 \times \text{NormCommission}) + (0.20 \times \text{NormExpectedReturn})$.
- [x] Unit tests verify that final Winning Score never exceeds 100 or drops below 0 across diverse mock distributions.

## Blocked by

- #3: Ticket 02: Baseline Snapshot & Sales Velocity Calculator

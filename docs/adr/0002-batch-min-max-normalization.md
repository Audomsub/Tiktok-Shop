# Batch Min-Max Normalization for Winning Score

The Go Analytics Engine normalizes Sales Velocity, Commission Rate, and Expected Return to a uniform 0–100 scale using batch Min-Max scaling with 99th percentile clipping before applying the 50/30/20 weights. We chose batch normalization over raw metric weighting to prevent monetary return or volume spikes from overpowering the percentage-based commission rate.

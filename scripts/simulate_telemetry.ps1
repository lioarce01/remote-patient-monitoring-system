# Configuración
$Url = "http://localhost:8081/observations"
$PatientId = "Patient123"
$Type = "heart_rate"
$Unit = "bpm"

# Obtener hora UTC como base
$baseTime = (Get-Date).ToUniversalTime()

Write-Host "⏱️ Enviando telemetrías normales..."
for ($i = 0; $i -lt 5; $i++) {
    $value = Get-Random -Minimum 77 -Maximum 80  # valores entre 77 y 79
    $timestamp = $baseTime.AddMinutes($i).ToString("yyyy-MM-ddTHH:mm:ssZ")

    $payload = @{
        patient_id = $PatientId
        type       = $Type
        value      = $value
        unit       = $Unit
        timestamp  = $timestamp
    } | ConvertTo-Json -Depth 3

    Write-Host "→ Normal: $value bpm @ $timestamp"

    Invoke-RestMethod -Uri $Url -Method Post -Body $payload -ContentType "application/json"
    Start-Sleep -Seconds 1
}

Write-Host "🚨 Enviando telemetría anómala..."
$anomalyValue = 95
$anomalyTimestamp = $baseTime.AddMinutes(5).ToString("yyyy-MM-ddTHH:mm:ssZ")

$anomalyPayload = @{
    patient_id = $PatientId
    type       = $Type
    value      = $anomalyValue
    unit       = $Unit
    timestamp  = $anomalyTimestamp
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri $Url -Method Post -Body $anomalyPayload -ContentType "application/json"
Write-Host "✅ Telemetrías enviadas. Verifica los logs o la base de datos para ver si se disparó una alerta."

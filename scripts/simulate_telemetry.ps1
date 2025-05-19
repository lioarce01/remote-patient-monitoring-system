# Configuración
$Url = "http://localhost:8081/observations"
$PatientId = "Patient123"
$Type = "heart_rate"
$Unit = "bpm"

# Obtener hora UTC como base
$baseTime = (Get-Date).ToUniversalTime()

Write-Host "Enviando telemetrias..."

for ($i = 0; $i -lt 30; $i++) {
    # Generar valores normales en rango típico 60-100 bpm, con algunas anómalas
    if ($i -eq 10 -or $i -eq 25) {
        # Anomalías: valores fuera del rango normal
        $value = Get-Random -Minimum 110 -Maximum 130
    } else {
        # Valores normales
        $value = Get-Random -Minimum 65 -Maximum 85
    }

    $timestamp = $baseTime.AddMinutes($i).ToString("yyyy-MM-ddTHH:mm:ssZ")

    $payload = @{
        patient_id = $PatientId
        type       = $Type
        value      = $value
        unit       = $Unit
        timestamp  = $timestamp
    } | ConvertTo-Json -Depth 3

    Write-Host "Telemetria: $value bpm @ $timestamp"

    Invoke-RestMethod -Uri $Url -Method Post -Body $payload -ContentType "application/json"
    Start-Sleep -Seconds 1
}

Write-Host "Telemetrias enviadas. Revisa los logs para ver las alertas."

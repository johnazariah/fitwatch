module FitSync.Config

open System
open System.IO
open System.Text.Json

/// Configuration stored in ~/.fitsync/config.json
type Config = {
    mywhooshToken: string option
    mywhooshWhooshId: string option
    intervalsApiKey: string option
    intervalsAthleteId: string option
    trainingPeaksToken: string option
    trainingPeaksAthleteId: int option
    igpsportToken: string option
    zwiftToken: string option
    lastSync: DateTimeOffset option
}

module Config =
    let private configDir = 
        Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".fitsync")
    
    let private configPath = Path.Combine(configDir, "config.json")
    
    let private jsonOptions = 
        let opts = JsonSerializerOptions(WriteIndented = true)
        opts.PropertyNamingPolicy <- JsonNamingPolicy.CamelCase
        opts
    
    let empty = {
        mywhooshToken = None
        mywhooshWhooshId = None
        intervalsApiKey = None
        intervalsAthleteId = None
        trainingPeaksToken = None
        trainingPeaksAthleteId = None
        igpsportToken = None
        zwiftToken = None
        lastSync = None
    }
    
    let load () =
        try
            if File.Exists(configPath) then
                let json = File.ReadAllText(configPath)
                JsonSerializer.Deserialize<Config>(json, jsonOptions)
            else
                empty
        with _ -> empty
    
    let save (config: Config) =
        Directory.CreateDirectory(configDir) |> ignore
        let json = JsonSerializer.Serialize(config, jsonOptions)
        File.WriteAllText(configPath, json)
        config
    
    let update (f: Config -> Config) =
        load() |> f |> save
    
    let setToken token whooshId =
        update (fun c -> { c with mywhooshToken = Some token; mywhooshWhooshId = Some whooshId })
    
    let setIntervals apiKey athleteId =
        update (fun c -> { c with intervalsApiKey = Some apiKey; intervalsAthleteId = Some athleteId })
    
    let updateLastSync () =
        update (fun c -> { c with lastSync = Some DateTimeOffset.UtcNow })

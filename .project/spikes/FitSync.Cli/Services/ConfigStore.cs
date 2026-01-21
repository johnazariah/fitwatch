using System.Text.Json;

namespace FitSync.Cli.Services;

public class ConfigStore
{
    private readonly string _configPath;
    private Dictionary<string, string> _config;

    public ConfigStore()
    {
        var configDir = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.UserProfile),
            ".fitsync");
        
        Directory.CreateDirectory(configDir);
        _configPath = Path.Combine(configDir, "config.json");
        _config = Load();
    }

    public string? Get(string key)
    {
        return _config.TryGetValue(key, out var value) ? value : null;
    }

    public void Set(string key, string value)
    {
        _config[key] = value;
        Save();
        Console.WriteLine($"Set {key}");
    }

    public void List()
    {
        foreach (var kvp in _config)
        {
            var displayValue = kvp.Key.Contains("password") || kvp.Key.Contains("apikey") 
                ? "****" 
                : kvp.Value;
            Console.WriteLine($"{kvp.Key} = {displayValue}");
        }
    }

    private Dictionary<string, string> Load()
    {
        if (!File.Exists(_configPath))
            return new Dictionary<string, string>();

        var json = File.ReadAllText(_configPath);
        return JsonSerializer.Deserialize<Dictionary<string, string>>(json) 
            ?? new Dictionary<string, string>();
    }

    private void Save()
    {
        var json = JsonSerializer.Serialize(_config, new JsonSerializerOptions { WriteIndented = true });
        File.WriteAllText(_configPath, json);
    }
}

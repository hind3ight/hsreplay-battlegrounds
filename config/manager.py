"""Configuration management module."""

import os
import yaml
from pathlib import Path
from typing import Any, Dict, Optional


class ConfigManager:
    """Manages configuration loading, saving, and merging."""

    def __init__(self, config_path: Optional[str] = None):
        """Initialize config manager.
        
        Args:
            config_path: Path to config file. If None, uses default location.
        """
        self.base_dir = Path(__file__).parent
        self.default_config_path = self.base_dir / "default_config.yaml"
        self.config_path = Path(config_path) if config_path else self._get_config_path()
        self._config: Dict[str, Any] = {}
        self.load()

    def _get_config_path(self) -> Path:
        """Get user config path from environment or default."""
        config_dir = os.environ.get("OBSIDIAN_CONFIG_DIR", 
                                    os.path.expanduser("~/.config/obsidian"))
        return Path(config_dir) / "config.yaml"

    def load(self) -> None:
        """Load configuration from file, merging with defaults."""
        # Load defaults first
        self._config = self._load_yaml(self.default_config_path)
        
        # Load and merge user config if exists
        if self.config_path.exists():
            user_config = self._load_yaml(self.config_path)
            self._config = self._merge_config(self._config, user_config)

    def _load_yaml(self, path: Path) -> Dict[str, Any]:
        """Load YAML file."""
        try:
            with open(path, 'r', encoding='utf-8') as f:
                return yaml.safe_load(f) or {}
        except Exception as e:
            print(f"Warning: Failed to load {path}: {e}")
            return {}

    def _merge_config(self, default: Dict, user: Dict) -> Dict:
        """Deep merge user config into default."""
        result = default.copy()
        for key, value in user.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._merge_config(result[key], value)
            else:
                result[key] = value
        return result

    def save(self, path: Optional[str] = None) -> None:
        """Save current configuration to file."""
        save_path = Path(path) if path else self.config_path
        save_path.parent.mkdir(parents=True, exist_ok=True)
        with open(save_path, 'w', encoding='utf-8') as f:
            yaml.dump(self._config, f, default_flow_style=False, allow_unicode=True)

    def get(self, key: str, default: Any = None) -> Any:
        """Get config value by dot-notation key."""
        keys = key.split('.')
        value = self._config
        for k in keys:
            if isinstance(value, dict):
                value = value.get(k)
            else:
                return default
            if value is None:
                return default
        return value

    def set(self, key: str, value: Any) -> None:
        """Set config value by dot-notation key."""
        keys = key.split('.')
        config = self._config
        for k in keys[:-1]:
            if k not in config:
                config[k] = {}
            config = config[k]
        config[keys[-1]] = value

    @property
    def all(self) -> Dict[str, Any]:
        """Get all configuration as dict."""
        return self._config.copy()


# Global instance
_manager: Optional[ConfigManager] = None


def get_manager() -> ConfigManager:
    """Get global config manager instance."""
    global _manager
    if _manager is None:
        _manager = ConfigManager()
    return _manager

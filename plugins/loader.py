"""Plugin loading system using importlib."""

import importlib
import inspect
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Callable


class Plugin:
    """Base class for plugins."""

    name: str = "base_plugin"
    version: str = "1.0.0"

    def on_load(self) -> None:
        """Called when plugin is loaded."""
        pass

    def on_unload(self) -> None:
        """Called when plugin is unloaded."""
        pass

    def on_enable(self) -> None:
        """Called when plugin is enabled."""
        pass

    def on_disable(self) -> None:
        """Called when plugin is disabled."""
        pass


class PluginLoader:
    """Manages plugin discovery, loading, and lifecycle."""

    def __init__(self, plugin_dir: Optional[str] = None):
        """Initialize plugin loader.
        
        Args:
            plugin_dir: Directory to load plugins from. Defaults to ./plugins.
        """
        self.plugin_dir = Path(plugin_dir) if plugin_dir else Path(__file__).parent
        self._plugins: Dict[str, Plugin] = {}
        self._enabled: Dict[str, bool] = {}

    def discover_plugins(self) -> List[str]:
        """Discover available plugins in plugin directory."""
        plugins = []
        if not self.plugin_dir.exists():
            return plugins

        for path in self.plugin_dir.glob("*.py"):
            if path.stem.startswith("_"):
                continue
            plugins.append(path.stem)
        return plugins

    def load_plugin(self, name: str) -> bool:
        """Load a plugin by name.
        
        Args:
            name: Plugin module name.
            
        Returns:
            True if loaded successfully.
        """
        if name in self._plugins:
            return True

        try:
            # Import plugin module
            module = importlib.import_module(f"plugins.{name}")
            
            # Find plugin class (subclass of Plugin)
            plugin_class = None
            for item_name, item in inspect.getmembers(module):
                if (inspect.isclass(item) 
                    and issubclass(item, Plugin) 
                    and item is not Plugin):
                    plugin_class = item
                    break

            if plugin_class is None:
                print(f"No Plugin subclass found in {name}")
                return False

            # Instantiate and register
            plugin_instance = plugin_class()
            plugin_instance.name = name
            self._plugins[name] = plugin_instance
            self._enabled[name] = True
            
            # Call on_load hook
            plugin_instance.on_load()
            plugin_instance.on_enable()
            
            return True

        except Exception as e:
            print(f"Failed to load plugin {name}: {e}")
            return False

    def unload_plugin(self, name: str) -> bool:
        """Unload a plugin by name."""
        if name not in self._plugins:
            return False

        try:
            plugin = self._plugins[name]
            plugin.on_disable()
            plugin.on_unload()
            
            # Remove from sys.modules
            module_name = f"plugins.{name}"
            if module_name in sys.modules:
                del sys.modules[module_name]
            
            del self._plugins[name]
            del self._enabled[name]
            return True

        except Exception as e:
            print(f"Failed to unload plugin {name}: {e}")
            return False

    def enable_plugin(self, name: str) -> bool:
        """Enable a loaded plugin."""
        if name not in self._plugins:
            return False
        if self._enabled.get(name):
            return True
        
        try:
            self._plugins[name].on_enable()
            self._enabled[name] = True
            return True
        except Exception as e:
            print(f"Failed to enable plugin {name}: {e}")
            return False

    def disable_plugin(self, name: str) -> bool:
        """Disable a loaded plugin."""
        if name not in self._plugins:
            return False
        if not self._enabled.get(name):
            return True

        try:
            self._plugins[name].on_disable()
            self._enabled[name] = False
            return True
        except Exception as e:
            print(f"Failed to disable plugin {name}: {e}")
            return False

    def load_all(self) -> List[str]:
        """Discover and load all plugins.
        
        Returns:
            List of successfully loaded plugin names.
        """
        loaded = []
        for name in self.discover_plugins():
            if self.load_plugin(name):
                loaded.append(name)
        return loaded

    def get_plugin(self, name: str) -> Optional[Plugin]:
        """Get loaded plugin by name."""
        return self._plugins.get(name)

    def list_plugins(self) -> List[Dict[str, Any]]:
        """List all loaded plugins with status."""
        return [
            {
                "name": name,
                "version": plugin.version,
                "enabled": self._enabled.get(name, False)
            }
            for name, plugin in self._plugins.items()
        ]

    def register_hook(self, name: str, callback: Callable) -> None:
        """Register a callback for a plugin hook.
        
        Args:
            name: Hook name (e.g., 'on_parse', 'on_sync')
            callback: Callable to invoke
        """
        if not hasattr(self, '_hooks'):
            self._hooks: Dict[str, List[Callable]] = {}
        if name not in self._hooks:
            self._hooks[name] = []
        self._hooks[name].append(callback)

    def trigger_hook(self, name: str, *args, **kwargs) -> List[Any]:
        """Trigger all callbacks for a hook.
        
        Args:
            name: Hook name
            *args, **kwargs: Arguments to pass to callbacks
            
        Returns:
            List of callback results
        """
        results = []
        if hasattr(self, '_hooks') and name in self._hooks:
            for callback in self._hooks[name]:
                try:
                    results.append(callback(*args, **kwargs))
                except Exception as e:
                    print(f"Hook {name} callback failed: {e}")
        return results


# Global instance
_loader: Optional[PluginLoader] = None


def get_loader() -> PluginLoader:
    """Get global plugin loader instance."""
    global _loader
    if _loader is None:
        _loader = PluginLoader()
    return _loader

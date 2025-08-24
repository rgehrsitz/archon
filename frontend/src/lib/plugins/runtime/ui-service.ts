/**
 * UI Service Implementation
 * 
 * Provides UI integration services for plugins.
 */

import type { UI, Command, PanelDescriptor, ModalDescriptor } from '../api.js';
import { PluginError } from '../api.js';

/**
 * UI service implementation for plugin integration.
 */
export class UIService implements UI {
  private registeredCommands = new Map<string, Command>();
  private activePanels = new Map<string, any>();

  /**
   * Registers a command that can be invoked from the UI.
   */
  registerCommand(cmd: Command): void {
    if (this.registeredCommands.has(cmd.id)) {
      throw new PluginError(
        `Command with ID '${cmd.id}' is already registered`,
        'COMMAND_ALREADY_REGISTERED',
        'ui'
      );
    }

    this.registeredCommands.set(cmd.id, cmd);
    
    // TODO: Integrate with actual UI command system
    console.log(`Registered plugin command: ${cmd.id} - ${cmd.title}`);
  }

  /**
   * Shows a plugin panel in the UI.
   */
  showPanel(panel: PanelDescriptor): void {
    if (this.activePanels.has(panel.id)) {
      throw new PluginError(
        `Panel with ID '${panel.id}' is already active`,
        'PANEL_ALREADY_ACTIVE',
        'ui'
      );
    }

    this.activePanels.set(panel.id, panel);
    
    // TODO: Integrate with actual UI panel system
    console.log(`Showing plugin panel: ${panel.id} - ${panel.title}`);
  }

  /**
   * Shows a modal dialog.
   */
  async showModal(modal: ModalDescriptor): Promise<void> {
    // TODO: Integrate with actual modal system
    console.log(`Showing plugin modal: ${modal.id} - ${modal.title}`);
    
    return new Promise(resolve => {
      // Simulate modal interaction
      setTimeout(resolve, 100);
    });
  }

  /**
   * Shows a notification toast.
   */
  notify(opts: { level: 'info' | 'warn' | 'error', message: string }): void {
    // TODO: Integrate with actual notification system (e.g., sonner)
    console.log(`Plugin notification [${opts.level}]: ${opts.message}`);
    
    // For now, use browser notifications as fallback
    switch (opts.level) {
      case 'info':
        console.info(`Plugin: ${opts.message}`);
        break;
      case 'warn':
        console.warn(`Plugin: ${opts.message}`);
        break;
      case 'error':
        console.error(`Plugin: ${opts.message}`);
        break;
    }
  }

  /**
   * Gets all registered commands.
   */
  getRegisteredCommands(): Command[] {
    return Array.from(this.registeredCommands.values());
  }

  /**
   * Executes a registered command by ID.
   */
  async executeCommand(commandId: string): Promise<void> {
    const command = this.registeredCommands.get(commandId);
    if (!command) {
      throw new PluginError(
        `Command '${commandId}' not found`,
        'COMMAND_NOT_FOUND',
        'ui'
      );
    }

    try {
      await command.execute();
    } catch (error) {
      throw new PluginError(
        `Command execution failed: ${error instanceof Error ? error.message : 'Unknown error'}`,
        'COMMAND_EXECUTION_FAILED',
        'ui'
      );
    }
  }

  /**
   * Closes a panel by ID.
   */
  closePanel(panelId: string): void {
    if (this.activePanels.has(panelId)) {
      this.activePanels.delete(panelId);
      console.log(`Closed plugin panel: ${panelId}`);
    }
  }

  /**
   * Gets all active panels.
   */
  getActivePanels(): PanelDescriptor[] {
    return Array.from(this.activePanels.values());
  }
}
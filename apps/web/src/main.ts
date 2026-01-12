import Alpine from 'alpinejs';
import '../styles/main.css';
import { registerGameStore } from './stores/game.js';
import { registerSongsStore } from './stores/songs.js';

// Initialize Alpine.js
declare global {
  interface Window {
    Alpine: typeof Alpine;
  }
}

window.Alpine = Alpine;

console.log('Bass Practice Game initializing...');

// Register stores before Alpine.start()
registerGameStore();
registerSongsStore();

// Start Alpine.js
Alpine.start();

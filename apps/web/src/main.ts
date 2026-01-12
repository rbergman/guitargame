import Alpine from 'alpinejs';
import '../styles/main.css';

// Initialize Alpine.js
declare global {
  interface Window {
    Alpine: typeof Alpine;
  }
}

window.Alpine = Alpine;

console.log('Bass Practice Game initializing...');

// Start Alpine.js
Alpine.start();

<script>
  import { onMount, onDestroy } from 'svelte';
  import { fade } from 'svelte/transition';
  import { EventsOn } from '../wailsjs/runtime/runtime';
  import { Start, Stop, Resume, Reset, GetLeft, GetStatus, SetDuration } from '../wailsjs/go/main/App';

  let timeLeft = 0;
  let status = 'stopped';
  let isPaused = false;
  let customHours = 0;
  let customMinutes = 5;
  let customSeconds = 0;
  let showInput = true;
  let audioElement = null;
  let unsubscribeTick = () => {};
  let unsubscribeFinished = () => {};

  // Константа для ключа в localStorage
  const STORAGE_KEY = 'timer_last_duration';

  function playSound() {
    if (audioElement && !audioElement.paused) return;
    audioElement = new Audio('/alert.mp3');
    audioElement.loop = true;
    audioElement.volume = 0.5;
    audioElement.play().catch(err => console.error("Ошибка звука:", err));
  }

  function stopSound() {
    if (audioElement) {
      audioElement.pause();
      audioElement.currentTime = 0;
      audioElement = null;
    }
  }

  // Функция отправки системного уведомления
  function sendSystemNotification() {
    if (!("Notification" in window)) return;

    if (Notification.permission === "granted") {
      new Notification("Таймер завершён!", {
        body: "Заданное время истекло.",
        icon: "/wails/build/appicon.png" // Путь к иконке вашего приложения Wails
      });
    }
  }

  onMount(async () => {
    // Запрашиваем права на отправку системных уведомлений при старте
    if ("Notification" in window && Notification.permission !== "granted") {
      Notification.requestPermission();
    }

    // Восстанавливаем последние сохраненные настройки времени
    const savedDuration = localStorage.getItem(STORAGE_KEY);
    if (savedDuration) {
      const totalSeconds = parseInt(savedDuration, 10);
      if (totalSeconds > 0) {
        customHours = Math.floor(totalSeconds / 3600);
        customMinutes = Math.floor((totalSeconds % 3600) / 60);
        customSeconds = totalSeconds % 60;
      }
    }

    timeLeft = await GetLeft();
    status = await GetStatus();
    isPaused = false;

    unsubscribeTick = EventsOn("tick", (seconds) => {
      timeLeft = seconds;
    });

    unsubscribeFinished = EventsOn("finished", () => {
      status = "finished";
      isPaused = false;
      playSound();
      sendSystemNotification(); // Вызов системного уведомления
    });
  });

  onDestroy(() => {
    if (unsubscribeTick) unsubscribeTick();
    if (unsubscribeFinished) unsubscribeFinished();
    stopSound();
  });

  async function handleStart() {
    try {
      await Start();
      status = "running";
      isPaused = false;
      stopSound();
    } catch (err) {
      alert("Ошибка: " + err);
    }
  }

  async function handleStop() {
    try {
      await Stop();
      status = "stopped";
      isPaused = true;
    } catch (err) {
      alert("Ошибка: " + err);
    }
  }

  async function handleResume() {
    try {
      await Resume();
      status = "running";
      isPaused = false;
    } catch (err) {
      alert("Ошибка: " + err);
    }
  }

  function handleReset() {
    Reset();
    status = "stopped";
    isPaused = false;
    showInput = true;
    timeLeft = 0;
    stopSound();
  }

  function applyTime(totalSeconds) {
    if (totalSeconds <= 0) {
      alert("Введите время больше 0!");
      return;
    }
    
    // Сохраняем выбранное время в localStorage
    localStorage.setItem(STORAGE_KEY, totalSeconds.toString());

    SetDuration(totalSeconds);
    timeLeft = totalSeconds;
    showInput = false;
    status = "stopped";
    isPaused = false;
  }

  function handleSetTime() {
    const totalSeconds = (customHours * 3600) + (customMinutes * 60) + customSeconds;
    applyTime(totalSeconds);
  }

  function handlePreset(minutes) {
    customHours = 0;
    customMinutes = minutes;
    customSeconds = 0;
    applyTime(minutes * 60);
  }

  function formatTime(seconds) {
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hrs > 0) {
      return `${String(hrs).padStart(2, '0')}:${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    }
    return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  }

  function getStatusLabel(currentStatus) {
    const labels = {
      'stopped': 'Готов к запуску',
      'running': 'Идет отсчет',
      'finished': 'Время вышло!'
    };
    return labels[currentStatus] || currentStatus;
  }
</script>

<main>
  {#if showInput}
    <div class="card" in:fade={{ duration: 200 }}>
      <h2>Установка времени</h2>
      
      <div class="time-inputs">
        <div class="input-group">
          <label for="hours">Часы</label>
          <input id="hours" type="number" bind:value={customHours} min="0" max="23" class="time-input" />
        </div>
        
        <span class="separator">:</span>

        <div class="input-group">
          <label for="minutes">Минуты</label>
          <input id="minutes" type="number" bind:value={customMinutes} min="0" max="59" class="time-input" />
        </div>
        
        <span class="separator">:</span>

        <div class="input-group">
          <label for="seconds">Секунды</label>
          <input id="seconds" type="number" bind:value={customSeconds} min="0" max="59" class="time-input" />
        </div>
      </div>

      <!-- Кнопки быстрого выбора (Пресеты) -->
      <div class="presets-container">
        <button class="btn btn-outline" on:click={() => handlePreset(5)}>5 мин</button>
        <button class="btn btn-outline" on:click={() => handlePreset(10)}>10 мин</button>
        <button class="btn btn-outline" on:click={() => handlePreset(25)}>25 мин (Pomodoro)</button>
      </div>
      
      <button class="btn btn-primary" on:click={handleSetTime}>
        Подтвердить
      </button>
    </div>
  {:else}
    <div class="card" class:pulse={status === 'finished'} in:fade={{ duration: 200 }}>
      <div class="timer-display">
        <h1 class="time">{formatTime(timeLeft)}</h1>
        <span class="status-badge status-{status}">
          {getStatusLabel(status)}
        </span>
      </div>

      <div class="controls">
        {#if status === 'stopped' && !isPaused}
          <button class="btn btn-success" on:click={handleStart}>Старт</button>
        {/if}

        {#if status === 'running'}
          <button class="btn btn-danger" on:click={handleStop}>Пауза</button>
        {/if}

        {#if status === 'stopped' && isPaused}
          <button class="btn btn-primary" on:click={handleResume}>Продолжить</button>
        {/if}

        <button class="btn btn-secondary" on:click={handleReset}>Сброс</button>
      </div>
    </div>
  {/if}
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    background-color: #f4f6f9;
    transition: background-color 0.3s ease;
  }

  @media (prefers-color-scheme: dark) {
    :global(body) {
      background-color: #0f172a;
    }
  }

  main {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 20px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    box-sizing: border-box;
  }

  .card {
    background: #ffffff;
    padding: 40px;
    border-radius: 24px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
    text-align: center;
    width: 100%;
    max-width: 420px;
    transition: background-color 0.3s ease, transform 0.3s ease;
  }

  @media (prefers-color-scheme: dark) {
    .card {
      background: #1e293b;
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
    }
  }

  h2 {
    margin: 0 0 32px 0;
    font-size: 1.5rem;
    font-weight: 600;
    color: #1e293b;
  }

  @media (prefers-color-scheme: dark) {
    h2 { color: #f8fafc; }
  }

  .time-inputs {
    display: flex;
    gap: 8px;
    justify-content: center;
    align-items: center;
    margin-bottom: 24px;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .input-group label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #64748b;
  }

  .separator {
    font-size: 2rem;
    font-weight: 300;
    color: #cbd5e1;
    margin-top: 20px;
  }

  @media (prefers-color-scheme: dark) {
    .separator { color: #475569; }
  }

  .time-input {
    width: 80px;
    padding: 16px 0;
    font-size: 2rem;
    font-weight: 500;
    text-align: center;
    border: 2px solid #e2e8f0;
    border-radius: 16px;
    background: transparent;
    color: #1e293b;
    outline: none;
    transition: all 0.2s ease;
  }

  .time-input:focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.15);
  }

  @media (prefers-color-scheme: dark) {
    .time-input {
      border-color: #334155;
      color: #f8fafc;
    }
    .time-input:focus {
      border-color: #3b82f6;
      box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.25);
    }
  }

  .time-input::-webkit-outer-spin-button,
  .time-input::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  /* Стили для блока пресетов */
  .presets-container {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: center;
    margin-bottom: 32px;
  }

  .presets-container .btn-outline {
    flex: 1 1 calc(50% - 8px);
    padding: 10px 14px;
    font-size: 0.9rem;
  }

  .presets-container .btn-outline:last-child {
    flex: 1 1 100%; /* Кнопка Помодоро растягивается во всю ширину */
  }

  .timer-display {
    margin-bottom: 40px;
  }

  .time {
    font-size: 5.5rem;
    font-weight: 200;
    margin: 0 0 16px 0;
    font-variant-numeric: tabular-nums;
    color: #0f172a;
    letter-spacing: -0.02em;
    line-height: 1;
  }

  @media (prefers-color-scheme: dark) {
    .time { color: #f8fafc; }
  }

  .status-badge {
    display: inline-block;
    padding: 6px 16px;
    border-radius: 100px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .status-stopped { background: #e2e8f0; color: #475569; }
  .status-running { background: #dbeafe; color: #1e40af; }
  .status-finished { background: #fee2e2; color: #991b1b; }

  @media (prefers-color-scheme: dark) {
    .status-stopped { background: #334155; color: #94a3b8; }
    .status-running { background: #1e3a8a; color: #93c5fd; }
    .status-finished { background: #7f1d1d; color: #fca5a5; }
  }

  .controls {
    display: flex;
    gap: 12px;
  }

  .controls .btn {
    flex: 1;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 16px 24px;
    font-size: 1rem;
    font-weight: 600;
    border: none;
    border-radius: 16px;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    box-sizing: border-box;
  }

  .btn:active {
    transform: scale(0.98);
  }

  .btn-primary { background: #3b82f6; color: white; }
  .btn-primary:hover { background: #2563eb; }

  .btn-success { background: #10b981; color: white; }
  .btn-success:hover { background: #059669; }

  .btn-danger { background: #ef4444; color: white; }
  .btn-danger:hover { background: #dc2626; }

  .btn-secondary { background: #64748b; color: white; }
  .btn-secondary:hover { background: #475569; }

  .btn-outline {
    background: transparent;
    border: 2px solid #e2e8f0;
    color: #475569;
  }
  .btn-outline:hover {
    background: #f1f5f9;
    border-color: #cbd5e1;
  }

  @media (prefers-color-scheme: dark) {
    .btn-secondary { background: #475569; }
    .btn-secondary:hover { background: #334155; }
    
    .btn-outline {
      border-color: #334155;
      color: #94a3b8;
    }
    .btn-outline:hover {
      background: #334155;
      color: #f8fafc;
    }
  }

  .pulse {
    animation: cardPulse 1.5s infinite;
  }

  @keyframes cardPulse {
    0% { box-shadow: 0 10px 30px rgba(239, 68, 68, 0.1); }
    50% { box-shadow: 0 10px 40px rgba(239, 68, 68, 0.4); }
    100% { box-shadow: 0 10px 30px rgba(239, 68, 68, 0.1); }
  }
</style>
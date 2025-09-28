import { useState, useEffect, useCallback, useRef } from 'react';
import styles from './Styles/game.module.css'

// --- Type Definitions ---
interface Position {
  x: number;
  y: number;
}

// --- Game Constants ---
const GAME_WIDTH = 500;
const GAME_HEIGHT = 600;
const PLAYER_SIZE = 30;
const PLAYER_STEP = 15;
const PROJECTILE_SIZE = 15;
const PROJECTILE_SPEED = 6;
const PROJECTILE_SPAWN_RATE = 300; // ms
const GAME_DURATION = 15; // seconds

// --- Game Component ---
export default function Game() {
  const [playerPos, setPlayerPos] = useState<Position>({ x: GAME_WIDTH / 2 - PLAYER_SIZE / 2, y: GAME_HEIGHT - PLAYER_SIZE - 10 });
  const playerPosRef = useRef<Position>(playerPos);

  const [projectiles, setProjectiles] = useState<Position[]>([]);
  const [timeLeft, setTimeLeft] = useState<number>(GAME_DURATION);
  const [gameOver, setGameOver] = useState<boolean>(false);
  const [gameWon, setGameWon] = useState<boolean>(false);
  const [gameStarted, setGameStarted] = useState<boolean>(false);

  const [petImage, setPetImage] = useState<string | null>(null);

  // Refs for intervals/animation frames to ensure they are cleared properly
  const gameLoopRef = useRef<number | null>(null);
  const projectileSpawnerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // --- Game Cleanup Logic ---
  const cleanupGame = useCallback(() => {
    if (gameLoopRef.current !== null) {
      cancelAnimationFrame(gameLoopRef.current);
      gameLoopRef.current = null;
    }
    if (projectileSpawnerRef.current !== null) {
      clearInterval(projectileSpawnerRef.current);
      projectileSpawnerRef.current = null;
    }
    if (timerRef.current !== null) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

   useEffect(() => {
    const savedPetImage = localStorage.getItem('pet');
    if (savedPetImage) {
      setPetImage(savedPetImage);
    }
  }, []);

  // --- Redirect on Game Over ---
  useEffect(() => {
    if (gameOver) {
      cleanupGame();
      setTimeout(() => {
        // In-app navigation would be preferable; simple redirect for this example.
        console.log("Redirecting to /");
        window.location.href = "/";
      }, 3000); // Wait 3 seconds to show the result
    }
  }, [gameOver, cleanupGame]);

  // --- Game Timer ---
  useEffect(() => {
    if (!gameStarted || gameOver) return;
    // clear any existing timer first
    if (timerRef.current !== null) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    timerRef.current = setInterval(() => {
      setTimeLeft(prevTime => {
        if (prevTime <= 1) {
          setGameOver(true);
          setGameWon(true);
          if (timerRef.current !== null) {
            clearInterval(timerRef.current);
            timerRef.current = null;
          }
          return 0;
        }
        return prevTime - 1;
      });
    }, 1000);
    return () => {
      if (timerRef.current !== null) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [gameStarted, gameOver]);

  // --- Safe player position setter that keeps ref in sync ---
  const setPlayerPosSafe = useCallback((updater: (prev: Position) => Position) => {
    setPlayerPos(prev => {
      const next = updater(prev);
      playerPosRef.current = next;
      return next;
    });
  }, []);

  // --- Player Movement ---
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (gameOver || !gameStarted) return;
    setPlayerPosSafe(prevPos => {
      let newX = prevPos.x;
      if (e.key === 'ArrowLeft' || e.key === 'a') {
        newX = Math.max(0, prevPos.x - PLAYER_STEP);
      } else if (e.key === 'ArrowRight' || e.key === 'd') {
        newX = Math.min(GAME_WIDTH - PLAYER_SIZE, prevPos.x + PLAYER_STEP);
      }
      return { ...prevPos, x: newX };
    });
  }, [gameOver, gameStarted, setPlayerPosSafe]);

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  // keep the ref in sync if playerPos is updated externally
  useEffect(() => {
    playerPosRef.current = playerPos;
  }, [playerPos]);

  // --- Main Game Loop (stable, reads player position from ref) ---
  const runGameLoop = useCallback(() => {
    if (gameOver) return;

    // 1. Move projectiles
    setProjectiles(prev =>
      prev
        .map(p => ({ ...p, y: p.y + PROJECTILE_SPEED }))
        .filter(p => p.y < GAME_HEIGHT)
    );

    // 2. Check for collisions using the ref
    setProjectiles(prevProjectiles => {
      for (let proj of prevProjectiles) {
        const pos = playerPosRef.current;
        const hit = (
          pos.x < proj.x + PROJECTILE_SIZE &&
          pos.x + PLAYER_SIZE > proj.x &&
          pos.y < proj.y + PROJECTILE_SIZE &&
          pos.y + PLAYER_SIZE > proj.y
        );
        if (hit) {
          setGameOver(true);
          setGameWon(false);
          break;
        }
      }
      return prevProjectiles;
    });

    gameLoopRef.current = requestAnimationFrame(runGameLoop);
  }, [gameOver]);

  // --- Start and manage game loop & projectile spawning ---
  useEffect(() => {
    if (gameStarted && !gameOver) {
      // Start game loop
      gameLoopRef.current = requestAnimationFrame(runGameLoop);

      // Start spawning projectiles
      if (projectileSpawnerRef.current !== null) {
        clearInterval(projectileSpawnerRef.current);
        projectileSpawnerRef.current = null;
      }
      projectileSpawnerRef.current = setInterval(() => {
        const newProjectile: Position = {
          x: Math.random() * (GAME_WIDTH - PROJECTILE_SIZE),
          y: -PROJECTILE_SIZE,
        };
        setProjectiles(prev => [...prev, newProjectile]);
      }, PROJECTILE_SPAWN_RATE);
    }

    // Cleanup function
    return () => {
      cleanupGame();
    };
    // intentionally excluding runGameLoop from deps to avoid restarting on every render;
    // runGameLoop is stable for our usage (reads current pos via ref)
  }, [gameStarted, gameOver, cleanupGame]);

  // --- Render Functions ---
 const StartScreen = () => (
    <div className={`${styles.overlay} ${styles.startOverlay}`}>
      <h2 className={styles.title}>DodgeFall</h2>
      <p className={styles.lead}>Dodge the falling blocks for {GAME_DURATION} seconds.</p>
      <p className={styles.leadSmall}>Use Arrow Keys or A/D to move.</p>

      <button
        type="button"
        onClick={() => {
          setProjectiles([]);
          setTimeLeft(GAME_DURATION);
          setGameOver(false);
          setGameWon(false);
          setGameStarted(true);
        }}
        className={styles.startButton}
      >
        Start Game
      </button>
    </div>
  );

  const GameOverScreen = () => (
    <div className={`${styles.overlay} ${styles.gameOverOverlay}`}>
      <h2 className={`${styles.gameOverTitle} ${gameWon ? styles.win : styles.lose}`}>
        {gameWon ? "You Survived!" : "Game Over"}
      </h2>
      <p className={styles.lead}>Redirecting back home...</p>
    </div>
  );

  return (
    <div className={styles.pageCenter}>
      <div
        className={styles.container}
        style={{ width: GAME_WIDTH, height: GAME_HEIGHT }}
      >
        {!gameStarted && <StartScreen />}
        {gameOver && <GameOverScreen />}

        <div className={styles.header}>
          <span className={styles.headerTitle}>DodgeFall</span>
          <span className={styles.headerTime}>Time Left: <span className={styles.timeValue}>{timeLeft}</span></span>
        </div>

        {/* Player: Now uses the pet image if available */}
        <div
          className={styles.player}
          style={{
            left: playerPos.x,
            top: playerPos.y,
            width: PLAYER_SIZE,
            height: PLAYER_SIZE,
            backgroundImage: petImage ? `url(${petImage})` : 'none',
            backgroundColor: petImage ? 'transparent' : '#67e8f9',
          }}
        />

        {projectiles.map((proj, i) => (
          <div
            key={i}
            className={styles.projectile}
            style={{
              left: proj.x,
              top: proj.y,
              width: PROJECTILE_SIZE,
              height: PROJECTILE_SIZE,
            }}
          />
        ))}
      </div>
    </div>
  );
}
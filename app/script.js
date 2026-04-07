/**
 * @fileoverview Enterprise Registration Portal Core Logic
 * @version 9.4.2-RELEASE-CANDIDATE-FINAL-v3
 * @author Enterprise Javascript Sub-committee on Redundancy
 * 
 * @description
 * This file contains the highly optimized, synergized, and robust JavaScript
 * required to execute the Enterprise Registration Portal. It leverages the
 * Document Object Model (DOM) to provide a rich, interactive, and actively
 * hostile user experience.
 * 
 * @security
 * This file is protected by multiple layers of imaginary encryption.
 * Do not look directly at the code without protective eyewear.
 */

document.addEventListener("DOMContentLoaded", () => {
    /**
     * DOM Element Caching Phase
     * We cache these elements once to save approximately 0.0001ms of processing time,
     * ensuring our application is blazing fast before we intentionally slow it down.
     */
    const terminal = document.getElementById('terminal-output');
    const bootScreen = document.getElementById('boot-screen');
    const errorState = document.getElementById('error-state');
    const app = document.getElementById('app');
    const cookiesOverlay = document.getElementById('cookies-overlay');
    
    /**
     * Theme Toggle Mechanism
     * Completely destroys the UI by turning it pitch black. 
     * Added because a VP read an article about "Dark Mode" on an airplane.
     */
    const themeToggle = document.getElementById('theme-toggle');
    themeToggle.addEventListener('click', (e) => {
        e.preventDefault(); // Prevents default behavior just in case
        document.body.classList.toggle('pitch-black'); // The nuclear option
    });

    /**
     * Fake Package Dependency Array
     * A curated list of 30 dependencies to make our project look complex.
     */
    const packages = ['react', 'vue', 'angular', 'jquery', 'lodash', 'express', 'webpack', 'babel', 'typescript', 'eslint', 'jest', 'rxjs', 'redux', 'graphql', 'socket.io', 'mongoose', 'passport', 'cors', 'helmet', 'morgan', 'chalk', 'inquirer', 'left-pad', 'is-even', 'enterprise-spaghetti-generator', 'aws-sdk', 'firebase', 'tailwindcss', 'bootstrap', 'material-ui'];
    
    /**
     * Boot Sequence Generation
     * Dynamically builds an array of string outputs to simulate a 
     * massive, brittle compilation process.
     */
    const bootSequence = [
        "INIT SYSTEM v9.4.2...",
        "LOADING KERNEL MODULES..."
    ];
    
    // Inject 150 random package installations to assert dominance
    for(let i=0; i<150; i++) {
        bootSequence.push(`npm install ${packages[Math.floor(Math.random() * packages.length)]}@latest...`);
    }
    
    // The inevitable conclusion of modern web development
    bootSequence.push(
        "Resolving dependencies (14,302 found)...",
        "Building native extensions...",
        "ERROR: MEMORY LEAK DETECTED IN is-even",
        "ERROR: left-pad returned undefined for padding length -1",
        "FATAL_CRASH"
    );

    let bootIndex = 0;
    let bootInterval;

    /**
     * @function startBoot
     * @description Initiates the fake terminal boot sequence. Prints lines at a rapid pace
     * before inevitably crashing and demanding user intervention.
     */
    function startBoot() {
        terminal.innerHTML = "";
        errorState.classList.add("hidden");
        bootIndex = 0;
        
        bootInterval = setInterval(() => {
            if (bootIndex < bootSequence.length) {
                const line = bootSequence[bootIndex];
                terminal.innerHTML += line + "<br>";
                terminal.scrollTop = terminal.scrollHeight; // Auto-scroll to bottom
                
                // When we hit the crash string, stop the interval and show the error state
                if (line === "FATAL_CRASH") {
                    clearInterval(bootInterval);
                    errorState.classList.remove("hidden");
                }
                bootIndex++;
            }
        }, 15); // 15ms delay ensures Maximum Hacking Speed visual effect
    }

    // Immediately invoke the boot sequence on load
    startBoot();

    /**
     * ======================================================================
     * LEVER MECHANISM LOGIC
     * ======================================================================
     */
    const leverHandle = document.getElementById('lever-handle');
    const leverTrack = document.getElementById('lever-track');
    const loadingContainer = document.getElementById('loading-container');
    const loadingBar = document.getElementById('loading-bar');
    
    let isDragging = false;
    let startY, startTop;

    // Listen for the initial grasp of the lever
    leverHandle.addEventListener('mousedown', (e) => {
        isDragging = true;
        startY = e.clientY;
        startTop = parseInt(window.getComputedStyle(leverHandle).top) || 0;
    });

    // Listen for the struggle as the user drags the lever down
    document.addEventListener('mousemove', (e) => {
        if (!isDragging) return;
        
        let dy = e.clientY - startY;
        let newTop = startTop + dy;
        let maxTop = leverTrack.clientHeight - leverHandle.clientHeight;
        
        // Boundaries
        if (newTop < 0) newTop = 0;
        
        // If they pull it all the way down, trigger the restart
        if (newTop >= maxTop) {
            newTop = maxTop;
            isDragging = false;
            leverHandle.style.top = newTop + 'px';
            triggerRestart();
        } else {
            // Otherwise, keep dragging
            leverHandle.style.top = newTop + 'px';
        }
    });

    // If they let go before hitting the bottom, snap it back to the top
    document.addEventListener('mouseup', () => {
        if (isDragging) {
            isDragging = false;
            leverHandle.style.transition = 'top 0.2s';
            leverHandle.style.top = '0px'; // The punishment
            setTimeout(() => { leverHandle.style.transition = ''; }, 200);
        }
    });

    /**
     * @function triggerRestart
     * @description Fires when the lever is fully pulled. Initiates the broken loading bar.
     */
    function triggerRestart() {
        document.getElementById('lever-container').style.display = 'none';
        loadingContainer.classList.remove('hidden');
        
        let progress = 0;
        let loadInterval = setInterval(() => {
            // Randomly jump the progress up
            progress += Math.floor(Math.random() * 20);
            loadingBar.style.width = progress + '%';
            loadingBar.innerText = progress + '%';
            
            // Allow it to overflow up to 450% to visually break the page
            if (progress > 450) {
                clearInterval(loadInterval);
                setTimeout(() => {
                    bootScreen.classList.add("hidden");
                    app.classList.remove("hidden");
                    // Immediately assault them with the cookies popup
                    cookiesOverlay.classList.remove('hidden');
                    startAnnoyingFeatures(); // Initialize the core features
                }, 1000);
            }
        }, 150);
    }

    /**
     * ======================================================================
     * COOKIE MODAL & HYDRA LOGIC
     * ======================================================================
     */
    const acceptCookiesBtn = document.getElementById('accept-cookies');
    const rejectCookiesBtn = document.getElementById('reject-cookies');
    const cookiePrefs = document.getElementById('cookie-preferences');
    const closeCookiesBtn = document.getElementById('close-cookies');
    const cookiesModal = document.getElementById('cookies-modal');

    // "Accept All" just threatens the user
    acceptCookiesBtn.addEventListener('click', (e) => {
        e.preventDefault();
        alert("Thanks! We are currently uploading your search history. Please wait...");
    });

    // "Rejecting" just opens a fake preferences menu with disabled checkboxes
    rejectCookiesBtn.addEventListener('click', (e) => {
        e.preventDefault();
        cookiePrefs.classList.remove('hidden');
        rejectCookiesBtn.innerText = "Manage Preferences ⮝";
    });

    // Closing the modal triggers the 12-second agonizing float-away animation
    closeCookiesBtn.addEventListener('click', (e) => {
        e.preventDefault();
        cookiesModal.classList.add('modal-closing'); // Applies the CSS transition
        
        // Spawn 2 new hydra modals before the first one is even gone
        setTimeout(() => {
            spawnHydraModal();
            spawnHydraModal();
        }, 1500);

        // Finally hide the overlay after 12 entire seconds
        setTimeout(() => {
            cookiesOverlay.classList.add('hidden');
        }, 12000);
    });

    /**
     * @function spawnHydraModal
     * @description Creates a dynamically injected DOM node that mimics a modal. 
     * If closed, it recursively spawns two more of itself.
     */
    function spawnHydraModal() {
        const hydra = document.createElement('div');
        hydra.className = 'hydra-modal vibe-card';
        hydra.innerHTML = `
            <h3>Are you sure?</h3>
            <p>Closing the previous modal has been logged.</p>
            <button class="btn-tertiary close-hydra">Close this too</button>
        `;
        
        // Randomly position the hydra somewhere on the screen
        const maxX = window.innerWidth - 350;
        const maxY = window.innerHeight - 200;
        hydra.style.left = Math.max(10, Math.floor(Math.random() * maxX)) + 'px';
        hydra.style.top = Math.max(10, Math.floor(Math.random() * maxY)) + 'px';
        
        document.body.appendChild(hydra);

        const closeBtn = hydra.querySelector('.close-hydra');
        
        // If they miraculously click it, two more take its place
        closeBtn.addEventListener('click', (e) => {
            e.preventDefault();
            spawnHydraModal();
            spawnHydraModal();
            hydra.remove(); // Kill the current head
        });
    }

    /**
     * ======================================================================
     * CORE APPLICATION FEATURES (THE GASLIGHTING)
     * ======================================================================
     */
    function startAnnoyingFeatures() {
        /**
         * 1. The Gaslighting Inputs
         * Intercepts keystrokes in specific input fields and randomly swaps
         * vowels out for other vowels to make the user think they are going crazy.
         */
        const inputs = document.querySelectorAll('.gaslight-input');
        const vowels = ['a', 'e', 'i', 'o', 'u'];
        
        inputs.forEach(input => {
            input.addEventListener('input', (e) => {
                let val = e.target.value;
                if (val.length === 0) return;
                
                const lastChar = val[val.length - 1];
                const lowerChar = lastChar.toLowerCase();
                
                // 40% chance to gaslight a vowel
                if (vowels.includes(lowerChar) && Math.random() > 0.6) {
                    const otherVowels = vowels.filter(v => v !== lowerChar);
                    const randomVowel = otherVowels[Math.floor(Math.random() * otherVowels.length)];
                    
                    // Preserve the casing of the original input
                    const replacement = lastChar === lastChar.toUpperCase() ? randomVowel.toUpperCase() : randomVowel;
                    
                    // Replace the character silently
                    e.target.value = val.slice(0, -1) + replacement;
                }
            });
        });

        /**
         * 2. The Uncontrollable Age Slider
         * Constantly updates the value of the range slider every 85ms.
         */
        const ageSlider = document.getElementById('age-slider');
        const ageDisplay = document.getElementById('age-display');
        const lockAgeBtn = document.getElementById('lock-age');
        
        let sliderInterval = setInterval(() => {
            const randomVal = Math.floor(Math.random() * 101);
            ageSlider.value = randomVal;
            ageDisplay.innerText = randomVal;
        }, 85);

        // Allow manual input, but it gets overwritten almost instantly
        ageSlider.addEventListener('input', (e) => {
            ageDisplay.innerText = e.target.value;
        });

        // The betrayal mechanic
        lockAgeBtn.addEventListener('click', () => {
            clearInterval(sliderInterval);
            lockAgeBtn.innerText = "🔒 Locked!";
            lockAgeBtn.disabled = true;
            ageSlider.disabled = true;
            
            // THE BETRAYAL: Clear all textual inputs as a penalty for succeeding
            inputs.forEach(input => input.value = '');
            
            // Unlock it after 3.5 seconds and start the chaos again
            setTimeout(() => {
                alert("Security protocol expired. Age lock released for your safety.");
                lockAgeBtn.innerText = "🔒 Lock";
                lockAgeBtn.disabled = false;
                ageSlider.disabled = false;
                
                sliderInterval = setInterval(() => {
                    const randomVal = Math.floor(Math.random() * 101);
                    ageSlider.value = randomVal;
                    ageDisplay.innerText = randomVal;
                }, 85);
            }, 3500);
        });

        /**
         * 3. The Evasive Submit Button
         * Detaches the submit button from the DOM flow and teleports it around
         * the entire browser window using absolute coordinates.
         */
        const submitBtn = document.getElementById('submit-btn');
        
        submitBtn.addEventListener('mouseover', () => {
            if (submitBtn.style.position !== 'fixed') {
                submitBtn.style.position = 'fixed';
            }
            
            // Calculate maximum bounds so the button doesn't go off-screen
            const maxX = window.innerWidth - submitBtn.clientWidth - 20;
            const maxY = window.innerHeight - submitBtn.clientHeight - 20;
            
            // Generate random X and Y coordinates
            const randomX = Math.max(10, Math.floor(Math.random() * maxX));
            const randomY = Math.max(10, Math.floor(Math.random() * maxY));
            
            // Apply the new coordinates
            submitBtn.style.left = randomX + 'px';
            submitBtn.style.top = randomY + 'px';
        });
    }
});

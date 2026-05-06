# Hackathon Final Hours: 2-Person Parallel Work Breakdown

To maximize your remaining time and ensure no one is waiting on the other, here is how you should split the remaining tasks into two completely independent streams.

---

## 👨‍💻 Person A: The Infrastructure & Demo Master
*Focus: Ensuring the code runs flawlessly on stage without internet dependency.*

**1. Sandbox & Docker Polish**
- Ensure Docker Desktop is running smoothly and Alpine images are pre-pulled (`docker pull alpine`).
- Test `configs/SAFETY.yaml` by adding 3 distinct rules (`rm -rf`, `drop table`, `kubectl delete`) and verify the Gatekeeper catches them all.

**2. The Local Llama Fallback**
- The `Advisor Agent` in `pkg/advisor/advisor.go` is currently configured to use `localLlamaExplain` as the fallback if no `OPENAI_API_KEY` is set. 
- *Task:* If you actually have Ollama installed, you can quickly write a real `http.Post` to `http://localhost:11434/api/generate` in that function. If not, ensure the hardcoded fallback text looks perfectly realistic for the judges.

**3. Demo Rehearsal (The Terminal Runner)**
- You are driving the keyboard during the pitch. Practice the exact sequence:
  `go run cmd/pulse/main.go --web` -> `ls` (Allow) -> `rm -rf /` (Intercept) -> `e` (Explain) -> `N` (Reject).

---

## 🎨 Person B: The Storyteller & UI Polish
*Focus: The Pitch, the Presentation, and the Dashboard Aesthetics.*

**1. Slide Deck Creation (Independent of Code)**
- Build the 2-3 slides needed for the pitch.
- Slide 1: The News Headlines (KiranaPro, Microsoft Bangalore outage).
- Slide 2: The Architecture Diagram (copy the Mermaid chart from `README.md` and render it visually).
- Slide 3: The Tagline ("Pulse: The Multi-Agent Shield").

**2. Web Dashboard UI Tweaks**
- Open `static/index.html`. You can edit this file and refresh the browser instantly without recompiling Go!
- *Task:* Add some fake historical data to `audit.db` using DB Browser for SQLite (or just run a bunch of test commands in the terminal) so the dashboard looks full and impressive when presented.
- *Task:* If you want to change colors or add a logo to the UI, do it in the `<style>` block.

**3. Pitch Memorization**
- Take the `demo_script.md` artifact and practice reciting it. It needs to fit perfectly into the 2-minute window while Person A drives the keyboard.

---

## 🤝 The Merge Point (1 Hour Before Deadline)
1. Stop all coding and UI tweaks.
2. Run the full End-to-End demo together. Person B speaks while Person A types.
3. Ensure both the terminal and the `http://localhost:8080` dashboard are visible on the projector screen simultaneously (split screen is best!).

# MultiBlox 🚀

**MultiBlox** is a powerful, lightweight multi-instance launcher for Roblox on macOS. It allows you to create, manage, and run multiple independent Roblox instances with custom names and icons, bypassing the standard single-client limitation.



---

## ✨ Features
* **Infinite Instances:** Create as many separate Roblox clones as your Mac can handle.
* **Custom Icons:** Choose from a variety of built-in icons to distinguish your accounts.
* **Automatic Updates:** Detects when the master Roblox app is updated and prompts for repairs.
* **Clean UI:** Simple, dark-mode native interface built with Wails and Vite.
* **Security Built-in:** Automatically handles ad-hoc signing and attribute clearing to ensure clones actually run.

---

## 📥 Installation

Because MultiBlox is an independent project, you need to follow these steps to bypass macOS security restrictions:

1.  **Download:** Grab the latest `MultiBlox.zip` from the [Releases](https://github.com/iigordev/multiblox-launcher/releases) page.
2.  **Move to Applications:** Unzip the file and drag `MultiBlox.app` into your **/Applications** folder. 
    * *Note: The app expects to find Roblox at `/Applications/Roblox.app`.*
3.  **First Launch (Gatekeeper):**
    * Right-click `MultiBlox.app` and select **Open**.
    * A warning will appear saying Apple cannot check it for malicious software. Click **Open** again.
    * *This is only required for the very first launch.*

---

## 🛠 How to Use

1.  **Create:** Click the **"Create New Instance"** card.
2.  **Customize:** Enter a name (e.g., "Alt1") and pick an icon.
3.  **Launch:** Hover over your new instance and hit the **Play** button.
4.  **Manage:** Use the **Quit** button to kill a specific instance or **Right-Click** a card to delete it.

---

## ⚠️ Troubleshooting

### "App is Damaged" or "Cannot be Opened"
If macOS refuses to open the app even after right-clicking, open your terminal and run:
```bash
xattr -rd com.apple.quarantine /Applications/MultiBlox.app

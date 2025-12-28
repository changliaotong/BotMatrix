import sys
import os

# Add SDK path to sys.path
sdk_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "../sdk/python"))
sys.path.append(sdk_path)

from botmatrix import BotMatrixPlugin

app = BotMatrixPlugin()

@app.on_message()
def handle_message(ctx):
    text = ctx.event.get("payload", {}).get("text", "")
    if text.startswith("/pyecho "):
        content = text.replace("/pyecho ", "")
        ctx.reply(f"Python SDK Echo: {content}")

if __name__ == "__main__":
    app.run()

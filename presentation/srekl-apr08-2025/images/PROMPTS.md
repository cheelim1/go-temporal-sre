# Image Generation Prompts for Temporal Idempotency Presentation

Use these prompts with DALL-E or ChatGPT's image generation to create visual aids for your presentation.

## double-billing.png

**Prompt:**
"A split-screen technical illustration showing a payment system failure scenario. On the left side, show a customer's account being charged twice for the same transaction. On the right side, show a stressed SRE engineer looking at error logs and multiple transaction records with identical IDs. Use a red color scheme to highlight the problem area, and include visual elements like duplicate transaction IDs, timestamps showing similar times, and a declining account balance. Style should be clean, professional, and suitable for a technical presentation."

## temporal-logo.png

**Prompt:**
"The official Temporal logo on a clean background. The logo features a stylized infinity symbol in blue-green gradient colors. The image should be crisp, professional, and suitable for a technical presentation with ample whitespace around the logo."

*Note: Instead of generating this, you may want to download the official logo from [Temporal's website](https://temporal.io/media-kit) for proper branding.*

## workflow-deduplication.png

**Prompt:**
"A technical diagram illustrating Temporal's workflow deduplication mechanism. Show multiple client requests (at least 5) with the same WorkflowID all being routed to a single workflow execution. Use a funnel or filter visual metaphor where multiple incoming arrows with the same ID converge to a single execution path. Include visual elements like workflow IDs, RunIDs, and activity execution. Use blue and green colors to represent successful deduplication, with clean lines and a minimal technical style suitable for SRE engineers. Add small computer/server icons to represent clients and the Temporal server."

## idempotency-proof.png

**Prompt:**
"A technical screenshot mockup showing terminal output from a Temporal idempotency test. The image should show five different clients attempting to start workflows with the same ID, and all receiving the same RunID. Important lines should be highlighted in green, especially where it shows '5 clients received workflow handles' and 'Activity execution count: 1'. The terminal should have a dark background with green and white text, styled like a typical command line interface. Include visual indicators like checkmarks or success icons next to the key proof points."

## sre-visibility.png

**Prompt:**
"A dashboard-style visualization showing Temporal's execution visibility features that benefit SREs. Include panels showing: workflow execution history with timestamps, activity retry statistics, workflow search interface, and system health metrics. Use a dark theme with blue and green highlights for important data points. The style should be clean and professional, resembling modern observability tools like Grafana or Datadog. Include small graphs, timelines, and status indicators that would appeal to SRE engineers."

## questions.png

**Prompt:**
"A friendly, engaging illustration representing a Q&A session about Temporal and idempotency. Show a stylized character (representing a presenter) next to thought bubbles or question marks in different sizes. Include visual elements related to the presentation topic like workflow symbols, idempotency concepts, and Temporal's logo subtly in the background. Use a vibrant but professional color scheme with blues and greens. The image should be clean, slightly playful but still appropriate for a technical presentation to SRE engineers."

## Additional Tips

1. **Consistency**: Request that all images maintain a consistent color scheme (blues and greens matching Temporal's branding) for a cohesive presentation.

2. **Style**: Specify "technical illustration style" or "clean infographic style" to ensure the images are appropriate for an SRE audience.

3. **Text**: If you want text embedded in the images, explicitly mention it in the prompt. Otherwise, you can add text directly in your presentation software.

4. **Iterations**: Don't hesitate to refine prompts based on initial results. Often the second or third attempt with adjustments yields better results.

5. **Export Size**: Request larger resolution images (1024×1024 or higher) to ensure they look crisp when displayed during your presentation.

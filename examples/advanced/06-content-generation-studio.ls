# LangSpace Advanced Example: Content Generation Studio
# Multi-format content creation with quality assurance, personalization, and publishing.
#
# This example demonstrates:
# - Multi-agent content creation
# - Parallel content generation
# - A/B testing and optimization
# - Brand voice consistency
# - Multi-channel publishing
# - Content performance analytics

# ============================================================================
# CONFIGURATION
# ============================================================================

config {
  default_model: "claude-sonnet-4-20250514"

  providers: {
    anthropic: {
      api_key: env("ANTHROPIC_API_KEY")
    }
    openai: {
      api_key: env("OPENAI_API_KEY")
    }
  }

  # Content settings
  content: {
    default_language: "en"
    supported_languages: ["en", "es", "fr", "de", "ja"]
    max_generation_retries: 3
    plagiarism_threshold: 0.15
  }

  # Publishing channels
  channels: {
    blog: {
      url: env("CMS_URL")
      api_key: env("CMS_API_KEY")
    }
    social: {
      twitter: env("TWITTER_API_KEY")
      linkedin: env("LINKEDIN_API_KEY")
      instagram: env("INSTAGRAM_API_KEY")
    }
    email: {
      provider: "sendgrid"
      api_key: env("SENDGRID_API_KEY")
    }
  }
}

# ============================================================================
# BRAND AND STYLE GUIDELINES
# ============================================================================

file "brand-voice" {
  contents: ```
    # Brand Voice Guidelines

    ## Core Values
    - Innovation with purpose
    - Human-centered technology
    - Transparency and trust
    - Accessible expertise

    ## Tone Attributes
    - **Confident but not arrogant**: Share expertise without condescension
    - **Friendly but professional**: Approachable yet authoritative
    - **Clear but not simplistic**: Explain complexity accessibly
    - **Optimistic but realistic**: Inspire without overpromising

    ## Language Guidelines

    ### Do:
    - Use active voice
    - Start with the benefit
    - Use concrete examples
    - Include calls to action
    - Address the reader directly ("you")

    ### Don't:
    - Use jargon without explanation
    - Make absolute claims without evidence
    - Use passive voice excessively
    - Write walls of text
    - Use clickbait tactics

    ## Voice by Channel

    ### Blog Posts
    - Thought leadership focus
    - 1200-2000 words optimal
    - Include practical takeaways
    - Use subheadings every 200-300 words

    ### Social Media
    - Twitter: Punchy, insight-driven, 1-2 hashtags max
    - LinkedIn: Professional, longer-form, industry-focused
    - Instagram: Visual-first, inspirational, community-building

    ### Email
    - Subject lines: 6-10 words, value-focused
    - Preview text: Complement, don't repeat subject
    - Body: Scannable, single CTA per email

    ## Formatting Standards
    - Headlines: Title case
    - Subheadings: Sentence case
    - CTAs: Action verbs + benefit
    - Numbers: Use digits for 10+, words for <10
  ```
}

file "content-templates/blog-post" {
  contents: ```
    # Blog Post Template

    ## Structure
    1. **Hook** (50-100 words)
       - Surprising statistic or question
       - Relatable problem statement
       - Bold claim

    2. **Context** (100-150 words)
       - Why this matters now
       - Who should care
       - What's at stake

    3. **Main Content** (800-1500 words)
       - 3-5 key points
       - Each point: claim + evidence + example
       - Subheadings for scannability

    4. **Practical Application** (200-300 words)
       - Actionable steps
       - Templates or frameworks
       - Real-world examples

    5. **Conclusion** (100-150 words)
       - Summarize key insights
       - Future implications
       - Clear CTA

    ## SEO Requirements
    - Primary keyword in: title, first paragraph, 2-3 subheadings, conclusion
    - Meta description: 150-160 characters, include CTA
    - Alt text for all images
    - Internal links: 2-3 minimum
    - External links: 1-2 authoritative sources
  ```
}

file "content-templates/social-media" {
  contents: ```
    # Social Media Templates

    ## Twitter/X

    ### Thread Starter
    [Hook with curiosity gap]
    🧵 Let me explain...

    ### Insight Post
    Most people think [common belief].
    But [contrarian insight].
    Here's why: [brief explanation]

    ### Engagement Post
    [Question or hot take]
    Wrong answers only 👇

    ## LinkedIn

    ### Thought Leadership
    I've spent [time] studying [topic].
    Here are [number] things I wish I knew earlier:

    [Point 1 - most surprising]
    [Point 2 - most actionable]
    [Point 3 - most counterintuitive]

    What would you add?

    ### Story Post
    [Hook - the result or lesson]

    [Story in 3-5 paragraphs]

    [The lesson]
    [CTA or question]

    ## Instagram

    ### Carousel
    - Slide 1: Bold headline + visual hook
    - Slides 2-7: One tip per slide
    - Final slide: CTA + summary
    - Caption: Expand on the value, include 5-10 hashtags
  ```
}

# ============================================================================
# TOOLS
# ============================================================================

mcp "cms" {
  transport: "stdio"
  command: "npx"
  args: ["-y", "@company/mcp-cms"]
}

tool "check_plagiarism" {
  description: "Check content for plagiarism"

  parameters: {
    content: string required "Content to check"
    threshold: number optional 0.15 "Similarity threshold"
  }

  handler: http {
    method: "POST"
    url: env("PLAGIARISM_API_URL") + "/check"
    headers: {
      "Authorization": "Bearer " + env("PLAGIARISM_API_KEY")
    }
    body: params
  }
}

tool "seo_analyze" {
  description: "Analyze content for SEO optimization"

  parameters: {
    content: string required
    target_keyword: string required
    url: string optional
  }

  handler: http {
    method: "POST"
    url: env("SEO_API_URL") + "/analyze"
    body: params
  }
}

tool "readability_score" {
  description: "Calculate readability metrics"

  parameters: {
    content: string required
  }

  handler: http {
    method: "POST"
    url: env("READABILITY_API_URL") + "/score"
    body: params
  }
}

tool "generate_image" {
  description: "Generate an image using AI"

  parameters: {
    prompt: string required "Image generation prompt"
    style: string optional "photo, illustration, 3d, abstract"
    size: string optional "1024x1024"
  }

  handler: http {
    method: "POST"
    url: "https://api.openai.com/v1/images/generations"
    headers: {
      "Authorization": "Bearer " + env("OPENAI_API_KEY")
    }
    body: {
      model: "dall-e-3",
      prompt: params.prompt,
      size: params.size,
      style: params.style
    }
  }
}

tool "translate" {
  description: "Translate content to another language"

  parameters: {
    content: string required
    target_language: string required
    preserve_formatting: bool optional true
  }

  handler: http {
    method: "POST"
    url: env("TRANSLATION_API_URL") + "/translate"
    body: params
  }
}

tool "publish_cms" {
  description: "Publish content to CMS"

  parameters: {
    title: string required
    content: string required
    category: string required
    tags: array optional
    featured_image: string optional
    publish_date: string optional
    status: string optional "draft, scheduled, published"
  }

  handler: mcp("cms").create_post
}

tool "post_twitter" {
  description: "Post to Twitter/X"

  parameters: {
    text: string required
    media: array optional "Media URLs to attach"
    reply_to: string optional "Tweet ID to reply to"
  }

  handler: http {
    method: "POST"
    url: "https://api.twitter.com/2/tweets"
    headers: {
      "Authorization": "Bearer " + env("TWITTER_BEARER_TOKEN")
    }
    body: params
  }
}

tool "post_linkedin" {
  description: "Post to LinkedIn"

  parameters: {
    text: string required
    media: object optional
    visibility: string optional "public, connections"
  }

  handler: http {
    method: "POST"
    url: "https://api.linkedin.com/v2/ugcPosts"
    headers: {
      "Authorization": "Bearer " + env("LINKEDIN_ACCESS_TOKEN")
    }
    body: params
  }
}

tool "get_analytics" {
  description: "Get content performance analytics"

  parameters: {
    content_id: string required
    platform: string required "blog, twitter, linkedin, instagram"
    metrics: array optional ["views", "engagement", "shares", "conversions"]
  }

  handler: http {
    method: "GET"
    url: env("ANALYTICS_API_URL") + "/content/" + params.content_id
    query: params
  }
}

# ============================================================================
# CONTENT SCRIPTS
# ============================================================================

# SEO keyword research
script "keyword-research" {
  language: "python"
  runtime: "python3"

  capabilities: [network]

  parameters: {
    topic: string required
    competitors: array optional
    region: string optional "us"
  }

  code: ```python
    import json
    import urllib.request
    import os

    # Simulate keyword research API call
    # In production, this would call SEMrush, Ahrefs, or similar
    url = f"{os.environ.get('SEO_API_URL', 'https://api.seo-tool.com')}/keywords"

    request_data = {
        "topic": topic,
        "competitors": competitors or [],
        "region": region or "us"
    }

    try:
        req = urllib.request.Request(
            url,
            data=json.dumps(request_data).encode(),
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {os.environ.get('SEO_API_KEY', '')}"
            },
            method="POST"
        )
        response = urllib.request.urlopen(req, timeout=30)
        data = json.loads(response.read())

        # Structure the output
        result = {
            "primary_keywords": data.get("keywords", [])[:5],
            "long_tail": data.get("long_tail", [])[:10],
            "questions": data.get("questions", [])[:5],
            "difficulty_scores": data.get("difficulty", {}),
            "search_volume": data.get("volume", {})
        }

    except Exception as e:
        # Fallback to basic analysis
        result = {
            "primary_keywords": [topic],
            "long_tail": [f"how to {topic}", f"best {topic}", f"{topic} guide"],
            "questions": [f"what is {topic}", f"why is {topic} important"],
            "error": str(e)
        }

    print(json.dumps(result, indent=2))
  ```
}

# Content analysis script
script "analyze-content" {
  language: "python"
  runtime: "python3"

  parameters: {
    content: string required
    brand_voice: string required
  }

  code: ```python
    import json
    import re
    from collections import Counter

    # Basic content analysis
    words = content.split()
    sentences = re.split(r'[.!?]+', content)

    # Readability (Flesch-Kincaid approximation)
    syllables = sum(len(re.findall(r'[aeiouAEIOU]', word)) for word in words)
    avg_sentence_length = len(words) / max(len(sentences), 1)
    avg_syllables = syllables / max(len(words), 1)
    flesch_score = 206.835 - 1.015 * avg_sentence_length - 84.6 * avg_syllables

    # Passive voice detection
    passive_patterns = [r'\b(is|are|was|were|be|been|being)\s+\w+ed\b']
    passive_count = sum(len(re.findall(p, content, re.I)) for p in passive_patterns)

    # Word frequency
    word_freq = Counter(word.lower() for word in words if len(word) > 4)

    # Sentiment indicators
    positive_words = ['great', 'excellent', 'innovative', 'powerful', 'easy']
    negative_words = ['difficult', 'complex', 'problem', 'issue', 'fail']

    pos_count = sum(1 for w in words if w.lower() in positive_words)
    neg_count = sum(1 for w in words if w.lower() in negative_words)

    result = {
        "word_count": len(words),
        "sentence_count": len(sentences),
        "avg_sentence_length": round(avg_sentence_length, 1),
        "readability": {
            "flesch_score": round(flesch_score, 1),
            "grade_level": "easy" if flesch_score > 60 else "moderate" if flesch_score > 30 else "difficult"
        },
        "style": {
            "passive_voice_instances": passive_count,
            "passive_voice_percentage": round(passive_count / max(len(sentences), 1) * 100, 1)
        },
        "sentiment": {
            "positive_words": pos_count,
            "negative_words": neg_count,
            "overall": "positive" if pos_count > neg_count else "negative" if neg_count > pos_count else "neutral"
        },
        "top_words": word_freq.most_common(10),
        "recommendations": []
    }

    # Generate recommendations
    if flesch_score < 50:
        result["recommendations"].append("Simplify language - aim for 8th grade reading level")
    if passive_count > len(sentences) * 0.2:
        result["recommendations"].append("Reduce passive voice usage")
    if avg_sentence_length > 25:
        result["recommendations"].append("Break up long sentences")
    if len(words) < 800:
        result["recommendations"].append("Consider expanding content for better SEO")

    print(json.dumps(result, indent=2))
  ```
}

# A/B variant generation
script "generate-variants" {
  language: "python"
  runtime: "python3"

  parameters: {
    content: object required "Original content"
    num_variants: number optional 3
    variation_type: string optional "headline, cta, tone"
  }

  code: ```python
    import json
    import random

    content = json.loads(content) if isinstance(content, str) else content
    num = num_variants or 3
    var_type = variation_type or "all"

    variants = []
    base = content

    # Headline variations
    headline_patterns = [
        "How to {verb} {topic} in {time}",
        "The Ultimate Guide to {topic}",
        "{number} Ways to {verb} {topic}",
        "Why {topic} Matters More Than Ever",
        "What Nobody Tells You About {topic}"
    ]

    # CTA variations
    cta_patterns = [
        "Get Started Now",
        "Learn More",
        "Try It Free",
        "See How It Works",
        "Start Your Journey"
    ]

    for i in range(num):
        variant = {
            "id": f"variant_{i+1}",
            "content": base.copy() if isinstance(base, dict) else {"text": base}
        }

        if var_type in ["headline", "all"]:
            variant["headline_variation"] = random.choice(headline_patterns)

        if var_type in ["cta", "all"]:
            variant["cta_variation"] = random.choice(cta_patterns)

        if var_type in ["tone", "all"]:
            tones = ["professional", "casual", "urgent", "inspirational"]
            variant["tone_variation"] = random.choice(tones)

        variants.append(variant)

    print(json.dumps({
        "original": content,
        "variants": variants,
        "test_recommendation": f"Test variants for {num * 7} days for statistical significance"
    }, indent=2))
  ```
}

# ============================================================================
# CONTENT AGENTS
# ============================================================================

agent "content-strategist" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.6

  instruction: ```
    You are a content strategist who plans content based on business goals and audience insights.

    Your responsibilities:
    1. Define content objectives and KPIs
    2. Identify target audience segments
    3. Research trending topics and keywords
    4. Create content calendars
    5. Determine optimal content formats

    When planning content:
    - Align with brand voice and values
    - Consider the buyer's journey stage
    - Balance evergreen and timely content
    - Plan for multi-channel distribution
    - Include measurement strategies

    Output a structured content brief including:
    - Objective and target audience
    - Key messages and angles
    - SEO keywords and topics
    - Content format and length
    - Distribution channels
    - Success metrics
  ```

  tools: [
    tool("seo_analyze"),
    tool("get_analytics"),
  ]

  scripts: [
    script("keyword-research")
  ]
}

agent "blog-writer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.7

  instruction: file("brand-voice") + "\n\n" + file("content-templates/blog-post")

  tools: [
    tool("seo_analyze"),
  ]
}

agent "social-media-writer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.8

  instruction: file("brand-voice") + "\n\n" + file("content-templates/social-media")
}

agent "content-editor" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.3

  instruction: ```
    You are a senior content editor responsible for quality and consistency.

    Review content for:
    1. **Brand Voice Alignment**
       - Tone and personality match
       - Consistent terminology
       - Value proposition clarity

    2. **Quality Standards**
       - Grammar and punctuation
       - Clarity and flow
       - Logical structure
       - Evidence and sources

    3. **Engagement Optimization**
       - Hook strength
       - Scanability
       - CTA effectiveness
       - Emotional resonance

    4. **Technical Requirements**
       - SEO optimization
       - Formatting standards
       - Length requirements
       - Platform specifications

    Provide:
    - Overall quality score (1-10)
    - Specific issues with line references
    - Concrete improvement suggestions
    - Final recommendation (approve/revise/reject)
  ```

  tools: [
    tool("readability_score"),
    tool("check_plagiarism"),
  ]

  scripts: [
    script("analyze-content")
  ]
}

agent "visual-designer" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.7

  instruction: ```
    You create visual concepts and generate images for content.

    For each piece of content:
    1. Identify key visual themes
    2. Create image prompts that align with brand
    3. Generate appropriate visuals
    4. Suggest image placement

    Image style guidelines:
    - Modern and clean aesthetic
    - Consistent color palette (blues, whites, subtle accents)
    - Human-centered when possible
    - Avoid stock photo clichés
    - Include alt text for accessibility
  ```

  tools: [
    tool("generate_image"),
  ]
}

agent "localization-specialist" {
  model: "claude-sonnet-4-20250514"
  temperature: 0.4

  instruction: ```
    You adapt content for different languages and cultures.

    Beyond translation:
    1. Cultural adaptation of examples and references
    2. Local market considerations
    3. Currency, date, and number formats
    4. Idiom and metaphor localization
    5. SEO optimization for local markets

    Maintain:
    - Brand voice consistency across languages
    - Technical accuracy
    - Formatting and structure
  ```

  tools: [
    tool("translate"),
  ]
}

# ============================================================================
# CONTENT PIPELINES
# ============================================================================

pipeline "create-blog-post" {
  # Step 1: Research and strategy
  step "strategize" {
    use: agent("content-strategist")

    input: {
      topic: $input.topic,
      audience: $input.audience,
      objectives: $input.objectives
    }
  }

  # Step 2: Generate content
  step "write" {
    use: agent("blog-writer")

    input: {
      brief: step("strategize").output,
      keywords: step("strategize").output.keywords
    }
  }

  # Step 3: Quality loop
  loop max: 3 {
    step "review" {
      use: agent("content-editor")

      input: {
        content: $current_draft,
        brand_voice: file("brand-voice"),
        brief: step("strategize").output
      }
    }

    break_if: step("review").output.recommendation == "approve"

    step "revise" {
      use: agent("blog-writer")

      input: {
        draft: $current_draft,
        feedback: step("review").output.feedback
      }
    }

    set $current_draft: step("revise").output
  }

  # Step 4: Visual creation
  step "create-visuals" {
    use: agent("visual-designer")

    input: step("write").output
  }

  # Step 5: Final assembly
  step "assemble" {
    input: {
      content: step("review").output.approved_content,
      images: step("create-visuals").output
    }

    output: {
      title: $content.title,
      body: $content.body,
      featured_image: $images.featured,
      meta_description: $content.meta_description,
      tags: step("strategize").output.tags
    }
  }

  output: step("assemble").output
}

pipeline "create-social-campaign" {
  # Step 1: Plan campaign
  step "plan" {
    use: agent("content-strategist")
    input: $input
  }

  # Step 2: Generate content for each platform in parallel
  step "generate" {
    parallel {
      step "twitter" {
        use: agent("social-media-writer")

        input: {
          brief: step("plan").output,
          platform: "twitter",
          format: "thread"
        }
      }

      step "linkedin" {
        use: agent("social-media-writer")

        input: {
          brief: step("plan").output,
          platform: "linkedin",
          format: "long-form"
        }
      }

      step "instagram" {
        use: agent("social-media-writer")

        input: {
          brief: step("plan").output,
          platform: "instagram",
          format: "carousel"
        }
      }
    }
  }

  # Step 3: Create visuals for each platform
  step "visuals" {
    parallel {
      step "twitter-image" {
        use: agent("visual-designer")
        input: { content: step("generate").twitter.output, platform: "twitter" }
      }

      step "linkedin-image" {
        use: agent("visual-designer")
        input: { content: step("generate").linkedin.output, platform: "linkedin" }
      }

      step "instagram-carousel" {
        use: agent("visual-designer")
        input: { content: step("generate").instagram.output, platform: "instagram" }
      }
    }
  }

  # Step 4: Generate A/B variants
  step "variants" {
    execute: script("generate-variants") {
      content: step("generate").output
      num_variants: 2
      variation_type: "headline"
    }
  }

  # Step 5: Review all content
  step "review" {
    use: agent("content-editor")

    input: {
      twitter: step("generate").twitter.output,
      linkedin: step("generate").linkedin.output,
      instagram: step("generate").instagram.output
    }
  }

  output: {
    twitter: {
      content: step("generate").twitter.output,
      image: step("visuals").twitter-image.output,
      variants: step("variants").output.variants
    },
    linkedin: {
      content: step("generate").linkedin.output,
      image: step("visuals").linkedin-image.output
    },
    instagram: {
      content: step("generate").instagram.output,
      carousel: step("visuals").instagram-carousel.output
    },
    schedule: step("plan").output.schedule
  }
}

pipeline "localize-content" {
  step "analyze-source" {
    execute: script("analyze-content") {
      content: $input.content
      brand_voice: file("brand-voice")
    }
  }

  step "translate" {
    parallel {
      for language in $input.languages {
        step "translate-{language}" {
          use: agent("localization-specialist")

          input: {
            content: $input.content,
            target_language: language,
            analysis: step("analyze-source").output
          }
        }
      }
    }
  }

  step "review-translations" {
    for language in $input.languages {
      use: agent("content-editor")

      input: {
        content: step("translate").output[language],
        language: language
      }
    }
  }

  output: step("review-translations").output
}

# ============================================================================
# PUBLISHING PIPELINES
# ============================================================================

pipeline "publish-blog" {
  step "prepare" {
    input: $input.content

    # Add final formatting
    output: {
      title: $content.title,
      body: $content.body,
      category: $input.category,
      tags: $content.tags,
      featured_image: $content.featured_image,
      status: $input.publish_now ? "published" : "scheduled",
      publish_date: $input.publish_date
    }
  }

  step "publish" {
    tools: [tool("publish_cms")]
    input: step("prepare").output
  }

  step "create-social" {
    run: pipeline("create-social-campaign") {
      input: {
        topic: $input.content.title,
        url: step("publish").output.url,
        type: "blog-promotion"
      }
    }
  }

  step "schedule-social" {
    parallel {
      step "twitter" {
        tools: [tool("post_twitter")]
        input: step("create-social").output.twitter
      }

      step "linkedin" {
        tools: [tool("post_linkedin")]
        input: step("create-social").output.linkedin
      }
    }
  }

  output: {
    blog_url: step("publish").output.url,
    social_posts: step("schedule-social").output
  }
}

# ============================================================================
# TRIGGERS
# ============================================================================

# Content request webhook
trigger "content-request" {
  event: http("/webhook/content-request") {
    method: "POST"
  }

  run: branch http.body.type {
    "blog" => pipeline("create-blog-post") {
      input: http.body
    }

    "social" => pipeline("create-social-campaign") {
      input: http.body
    }

    "localization" => pipeline("localize-content") {
      input: http.body
    }
  }

  on_complete: {
    slack.post(
      channel: "#content",
      text: "Content created: " + output.title,
      attachments: [{ text: output.preview }]
    )
  }
}

# Scheduled content generation
trigger "weekly-content" {
  event: schedule("0 9 * * MON")  # Every Monday at 9 AM

  run: {
    # Get trending topics
    trends: script("keyword-research") {
      topic: config.content.focus_area
    }

    # Generate content plan for the week
    plan: agent("content-strategist") {
      input: trends.output
    }

    # Create first piece
    pipeline("create-blog-post") {
      input: plan.output.posts[0]
    }
  }
}

# Performance analysis trigger
trigger "content-performance" {
  event: schedule("0 8 * * *")  # Daily at 8 AM

  run: {
    # Get performance data
    analytics: tool("get_analytics") {
      platform: "all"
      metrics: ["views", "engagement", "conversions"]
    }

    # Analyze and recommend
    agent("content-strategist") {
      input: analytics

      instruction: "Analyze content performance and recommend optimizations."
    }
  }

  on_complete: {
    email.send(
      to: "content-team@company.com",
      subject: "Daily Content Performance Report",
      body: output.report
    )
  }
}

# ============================================================================
# CLI ENTRYPOINTS
# ============================================================================

intent "create-blog" {
  params: {
    topic: string required "Blog topic"
    audience: string optional "Target audience"
    keywords: array optional "Target keywords"
    publish: bool optional false "Publish immediately"
  }

  run: pipeline("create-blog-post") {
    input: params
  }

  output: stdout
}

intent "create-social" {
  params: {
    topic: string required "Campaign topic"
    platforms: array optional ["twitter", "linkedin"] "Target platforms"
  }

  run: pipeline("create-social-campaign") {
    input: params
  }

  output: stdout
}

intent "translate" {
  params: {
    content_id: string required "Content ID to translate"
    languages: array required "Target languages"
  }

  run: pipeline("localize-content") {
    input: params
  }

  output: stdout
}

intent "analyze" {
  params: {
    content: string required "Content to analyze"
  }

  run: script("analyze-content") {
    content: params.content
    brand_voice: file("brand-voice")
  }

  output: stdout
}

intent "publish" {
  params: {
    content_id: string required "Content ID to publish"
    schedule: string optional "Publish schedule (ISO date or 'now')"
    channels: array optional ["blog", "twitter", "linkedin"]
  }

  run: pipeline("publish-blog") {
    input: {
      content_id: params.content_id,
      publish_now: params.schedule == "now",
      publish_date: params.schedule
    }
  }

  output: stdout
}

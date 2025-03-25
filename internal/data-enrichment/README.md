# README

## Project Overview
This project demonstrates a data enrichment application using Temporal.io's Go SDK, 
scheduled to run every 6 hours, ensuring only one instance runs at a time. 
It includes workflows for enriching customer data, activities for fetching and merging 
demographics, and comprehensive tests.

## Key Points
It seems likely that Temporal.io can be used for data enrichment by creating workflows 
to manage data processing steps.

Research suggests that Temporal.io’s durable execution and retry mechanisms make it 
suitable for handling complex, failure-prone enrichment tasks.

The evidence leans toward using Temporal.io for orchestrating steps like fetching, 
merging, and storing enriched data, ensuring reliability.

**What is Data Enrichment?**

Data enrichment involves enhancing existing data by adding more context or information from various sources, 
such as external APIs or databases, to make it more valuable for analysis. For example, adding demographic 
details to customer records can improve marketing strategies.

**How Temporal.io Fits In** 

Temporal.io is a platform designed for managing durable workflows, meaning it can handle long-running processes 
that need to be resilient to failures. It allows developers to define workflows in code, with the platform 
managing retries, state, and scaling. This makes it ideal for data enrichment, where processes like fetching
data from unreliable APIs or merging large datasets can fail.

**Using Temporal.io for Data Enrichment**

You can use Temporal.io to create a workflow that orchestrates the data enrichment process. 
This might include: Identifying the data to enrich, like customer lists.

    - Fetching additional data, such as demographic information, via activities.
    - Processing and merging this data to create enriched datasets.
    - Storing the final data in a warehouse or database.

Temporal.io ensures reliability by handling failures, such as retrying failed API calls, and 
provides tools for monitoring the workflow. This is particularly useful for large datasets, 
where parallel processing can be implemented using child workflows.

**Stretch Goal**

Leverage Codec to decrypt data on-the-fly when it actually needs to be used. Cryptographic material would be KMS;
that should be ready to be rotated.
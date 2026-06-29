# Requirements Document: BD Govt Job Circular Portal

## 1. Project Goal
The goal of this project is to build a centralized platform that automatically gathers and displays government job circulars in Bangladesh. (Non-Govt job circuler in future) It allows users to easily find, save, auto apply and get notified about relevant job opportunities.

## 2. Target Audience
- **Job Seekers:** People looking for government jobs in Bangladesh who want to search, filter, and track applications.
- **System Administrators:** Staff who manage the platform, monitor its operations, and sometimes manually update job listings.

## 3. Core Features

### For Job Seekers (Users)
- **User Accounts:** 
  - Users can create an account using their email.
  - They must verify their email to complete registration.
  - Secure login and the ability to reset a forgotten password.
  - Profile management (updating personal details like education level, district, and phone number).
- **Job Browsing and Searching:**
  - View a list of recent government jobs.
  - Search for jobs using keywords.
  - Filter jobs by category (e.g., BCS, Bank Jobs, Defense, Education), deadline, education requirements, and more.
  - View detailed information for a specific job (vacancies, age limits, salary, application instructions, required education/experience, and official circular documents).
- **Bookmarks:**
  - Save or bookmark jobs to view them later.
  - Add personal notes to bookmarked jobs.
- **Alerts and Notifications:**
  - Create customized job alerts based on specific keywords, categories, or organizations.
  - Receive automated notifications when a new job matches their alert preferences.

### For Administrators
- **Dashboard:** View overall statistics about the platform (e.g., number of users, active jobs).
- **User Management:** View the list of registered users.
- **Job Management:** 
  - Manually add, edit, or delete job postings.
  - Highlight or "feature" important job circulars so they appear prominently.
- **System Monitoring:** 
  - Manually start the process of fetching new jobs.
  - View logs and history of the automated job collection process to ensure it is running smoothly.

## 4. Automated System Processes
- **Automated Job Collection:** The system routinely checks various external sources (like job portals and government websites) to find and collect new job circulars without human intervention.
- **Duplicate Prevention:** The system checks new jobs against existing ones to ensure the same job is not listed multiple times.
- **Automatic Expiration:** Jobs that have passed their application deadline are automatically marked as expired and removed from the active job list.

## 5. Information Tracked by the System
- **User Profiles:** Basic contact and educational details.
- **Job Circulars:** Comprehensive details about each job (title, organization, dates, requirements, salary, application links, and attached images/PDFs).
- **Categories & Organizations:** A structured list of job categories and government organizations to help organize the jobs.

## 6. Implementation Checklist
- [x] Create simplest UI with Login and Registration.
- [x] Integrate Frontend into the Backend codebase (served via static files).
- [x] Skip email verification for instant registration and login.

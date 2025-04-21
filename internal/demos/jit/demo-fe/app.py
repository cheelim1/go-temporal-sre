#!/usr/bin/env python3

import streamlit as st
import requests
import json

# Configure the backend API base URL.
BACKEND_URL = "http://localhost:8080"

st.title("MongoDB Atlas JIT Access Request")

# Step 1: Get the list of existing database users
try:
    users_response = requests.get(f"{BACKEND_URL}/api/database-users")
    users_response.raise_for_status()
    database_users = users_response.json()
except Exception as e:
    st.error(f"Error loading database users: {e}")
    database_users = []

# Create a dropdown for username selection
username = st.selectbox("Select your database username:", database_users)

if st.button("Load User Info") and username:
    try:
        # Call the backend to get the current role.
        role_response = requests.get(f"{BACKEND_URL}/api/user-role", params={"username": username})
        role_response.raise_for_status()
        role_data = role_response.json()
        current_role = role_data.get("current_role", "")
        
        # Call the backend to get the list of built-in roles.
        roles_response = requests.get(f"{BACKEND_URL}/api/built-in-roles")
        roles_response.raise_for_status()
        built_in_roles = roles_response.json()
        
        st.session_state.current_role = current_role
        st.session_state.built_in_roles = built_in_roles
        st.success(f"Loaded info for user '{username}'. Current role: {current_role}")
    except Exception as e:
        st.error(f"Error loading user info: {e}")

# Only show the form if we have loaded the user info.
if "current_role" in st.session_state and "built_in_roles" in st.session_state:
    st.write(f"**Current Role:** {st.session_state.current_role}")
    
    # Filter out the current role from the list of built-in roles.
    available_roles = [r for r in st.session_state.built_in_roles if r != st.session_state.current_role]
    
    new_role = st.selectbox("Select the new role you want to request:", available_roles)
    reason = st.text_area("Reason for access (required):", placeholder="Please provide a reason for requesting this access")
    duration = st.selectbox("Select duration of access:", ["3m", "5m", "15m", "1h"])
    
    # Only disable submit button if reason is empty
    submit_disabled = not reason.strip()
    
    if st.button("Submit JIT Request", disabled=submit_disabled):
        if new_role == "atlasAdmin":
            st.error("Your user is not granted access to the atlasAdmin role.")
        elif not reason.strip():
            st.error("Please provide a reason for requesting access.")
        else:
            payload = {
                "username": username,
                "reason": reason,
                "new_role": new_role,
                "duration": duration,
            }
            try:
                response = requests.post(
                    f"{BACKEND_URL}/api/jit-request",
                    json=payload,
                    headers={"Content-Type": "application/json"},
                )
                response.raise_for_status()
                result = response.json()
                st.success(f"Request submitted successfully! Workflow ID: {result.get('workflowID')}, Run ID: {result.get('runID')}")
            except Exception as e:
                st.error(f"Failed to submit JIT request: {e}")

import streamlit as st
import requests
import json
from datetime import datetime
import time

# Configuration
API_BASE_URL = "http://localhost:8080"

def main():
    st.title("Breakglass Emergency Actions")
    st.write("Use this interface to execute emergency actions on services.")

    # Service selection
    service_id = st.text_input("Service ID", "service-123")

    # Action selection
    action = st.selectbox(
        "Action",
        ["restart", "scale", "rollback"],
        help="Select the emergency action to perform"
    )

    # Parameters
    st.subheader("Parameters")
    parameters = {}
    if action == "scale":
        parameters["replicas"] = st.number_input("Number of replicas", min_value=1, value=3)
    elif action == "rollback":
        parameters["version"] = st.text_input("Version to rollback to", "v1.0.0")

    # Requested by
    requested_by = st.text_input("Requested by", "admin@example.com")

    # Submit button
    if st.button("Execute Action"):
        # Prepare request
        data = {
            "service_id": service_id,
            "action": action,
            "parameters": parameters,
            "requested_by": requested_by
        }

        # Send request
        try:
            response = requests.post(
                f"{API_BASE_URL}/api/breakglass",
                json=data
            )
            response.raise_for_status()
            result = response.json()

            # Show success message
            st.success(f"Action started successfully! Workflow ID: {result['workflow_id']}")

            # Poll for status
            workflow_id = result['workflow_id']
            status = poll_workflow_status(workflow_id)
            
            if status:
                if status['success']:
                    st.success(f"Action completed successfully: {status['message']}")
                else:
                    st.error(f"Action failed: {status['message']}")
            else:
                st.warning("Could not get final status of the action")

        except requests.exceptions.RequestException as e:
            st.error(f"Failed to execute action: {str(e)}")

def poll_workflow_status(workflow_id):
    """Poll the workflow status until completion"""
    max_attempts = 10
    attempt = 0

    while attempt < max_attempts:
        try:
            response = requests.get(
                f"{API_BASE_URL}/api/breakglass/status",
                params={"workflow_id": workflow_id}
            )
            response.raise_for_status()
            status = response.json()

            if status['status'] == "COMPLETED":
                return status

            # Wait before next attempt
            time.sleep(2)
            attempt += 1

        except requests.exceptions.RequestException:
            return None

    return None

if __name__ == "__main__":
    main() 
import { NextRequest, NextResponse } from "next/server";

const ORCHESTRATOR_URL = process.env.ORCHESTRATOR_URL || "http://localhost:8080";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();

    // Map frontend config to backend format
    const backendPayload = {
      project_name: body.project_name,
      client_id: body.client_id || "ui-client",
      initial_message: body.initial_message || `Build a ${body.project_name} application`,
      enable_full_stack: body.enable_full_stack ?? true,
      enable_deployment: body.enable_deployment ?? true,
      enable_sre: body.enable_sre ?? true,
      preferred_frontend: body.preferred_frontend || "react",
      preferred_backend: body.preferred_backend || "nodejs",
      preferred_database: body.preferred_database || "postgresql",
      preferred_cloud: body.preferred_cloud || "aws",
    };

    console.log("Creating project:", backendPayload);

    const response = await fetch(`${ORCHESTRATOR_URL}/api/pods`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(backendPayload),
    });

    if (!response.ok) {
      const errorText = await response.text();
      console.error("Backend error:", errorText);
      return NextResponse.json(
        { error: "Failed to create project", details: errorText },
        { status: response.status }
      );
    }

    const data = await response.json();
    console.log("Project created:", data);

    return NextResponse.json(data);
  } catch (error) {
    console.error("Error creating project:", error);
    return NextResponse.json(
      { error: "Internal server error", details: String(error) },
      { status: 500 }
    );
  }
}

export async function GET() {
  try {
    const response = await fetch(`${ORCHESTRATOR_URL}/api/pods`);

    if (!response.ok) {
      return NextResponse.json(
        { error: "Failed to fetch projects" },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching projects:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}

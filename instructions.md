

1. Final Project Name
ParkIntel: AI-Driven Parking Hotspot and Enforcement Prioritization System


2. Final Problem Statement
Current enforcement against illegal parking is mostly patrol-based and reactive. Traffic police do not have a data-driven way to identify where illegal parking repeatedly occurs, when it is likely to happen, and which locations should be prioritized because of their potential congestion impact.
This project uses historical police violation records to:
1. Detect recurring illegal parking hotspots
2. Predict future hotspot risk
3. Estimate congestion-impact potential using violation severity
4. Rank enforcement zones for targeted action
5. Show hotspots on a dashboard map


3. Dataset Used
Only one dataset:
jan to may police violation_anonymized791b166.csv

Observed structure:
Rows: ~298,450
Columns: 24

Important columns:
id
latitude
longitude
location
vehicle_type
violation_type
offence_code
created_datetime
modified_datetime
police_station
junction_name
validation_status
validation_timestamp

Useful facts from the dataset:
WRONG PARKING: ~164k+
NO PARKING: ~139k+
PARKING IN A MAIN ROAD: ~23k+
DOUBLE PARKING: ~2k+
PARKING NEAR ROAD CROSSING: ~1.6k+
PARKING NEAR TRAFFIC LIGHT/ZEBRA: ~525

This is enough to build a strong prototype.

4. What the Model Can and Cannot Do
Can do
Predict illegal parking hotspot risk
Predict expected violations in a zone/time window
Detect repeat hotspots
Rank enforcement zones
Estimate congestion-impact potential using proxy scoring
Explain why a zone is high priority

Cannot honestly do
Predict actual traffic speed
Predict real traffic volume
Measure actual congestion reduction
Prove illegal parking caused congestion

Correct claim:
The system estimates likely congestion impact based on violation severity, location density, vehicle type, junction proximity, and repeat occurrence.


5. Final Product Requirement Document
Product Goal
Help traffic police move from reactive patrol-based enforcement to data-driven targeted enforcement.

Target Users
User
Need
Traffic control room
View city-level illegal parking hotspots
Field enforcement team
Get priority locations for patrol
Police station officer
Understand repeat hotspots in their jurisdiction
City planners
Analyze long-term violation patterns


Core User Stories
User Story 1: Hotspot map
As a traffic officer, I want to see illegal parking hotspots on a map so I can identify problem zones quickly.

User Story 2: Enforcement priority
As an enforcement lead, I want a ranked list of high-priority zones so I can deploy teams efficiently.

User Story 3: Time-based prediction
As a control room operator, I want to know when a hotspot is likely to become active again.

User Story 4: Explainability
As a decision-maker, I want to know why a zone is marked high priority.

User Story 5: Police-station filtering
As a station officer, I want to filter hotspots under my police station.


6. MVP Features
Must-have
Feature
Description
Hotspot heatmap
Show dense illegal parking zones
Risk prediction
Predict hotspot risk by zone/time
Congestion-impact proxy
Estimate likely carriageway obstruction impact
Enforcement ranking
Rank top zones for deployment
Explanation panel
Show reasons for each recommendation
Filters
Date, hour, police station, violation type, risk level
Zone-level analytics
Show repeat count, severity, vehicle mix


Good-to-have
Feature
Description
Patrol route suggestion
Suggest sequence of hotspots
Alert generation
Show zones crossing high-risk threshold
Before/after simulation
Simulate impact if violations reduce
Model confidence
Show high/medium/low confidence
Export report
Download station-wise hotspot report


7. Final System Architecture
                ┌──────────────────────────────┐
                 │ HackerEarth Police CSV        │
                 │ jan-may violation dataset     │
                 └───────────────┬──────────────┘
                                 │
                                 ▼
                 ┌──────────────────────────────┐
                 │ Data Cleaning Pipeline        │
                 │ - parse timestamps            │
                 │ - parse violation_type list   │
                 │ - clean coordinates           │
                 │ - handle missing values       │
                 └───────────────┬──────────────┘
                                 │
                                 ▼
                 ┌──────────────────────────────┐
                 │ Zone Generation Layer         │
                 │ - lat/lon grid or H3 cells    │
                 │ - assign each violation zone  │
                 └───────────────┬──────────────┘
                                 │
                                 ▼
                 ┌──────────────────────────────┐
                 │ Feature Engineering Layer     │
                 │ - hourly/daily aggregation    │
                 │ - violation counts            │
                 │ - vehicle severity            │
                 │ - junction risk               │
                 │ - repeat hotspot score        │
                 └───────────────┬──────────────┘
                                 │
                                 ▼
                 ┌──────────────────────────────┐
                 │ ML Layer                      │
                 │ Model 1: Hotspot prediction   │
                 │ Model 2: Risk classification  │
                 └───────────────┬──────────────┘
                                 │
                                 ▼
                 ┌──────────────────────────────┐
                 │ Scoring Layer                 │
                 │ - congestion proxy score      │
                 │ - enforcement priority score  │
                 │ - explanation generation      │
                 └───────────────┬──────────────┘
                                 │
                                 ▼
┌──────────────────────┐   ┌──────────────────────────────┐
│ Next.js Dashboard    │◀──│ Go Backend API                │
│ Map + ranking UI     │   │ REST APIs                     │
└──────────────────────┘   └──────────────────────────────┘


8. Recommended Tech Stack
Layer
Tech
Frontend
Next.js + TypeScript
Map
Leaflet / Mapbox visual layer
Backend
Go + Gin/Fiber
ML pipeline
Python + Pandas + Scikit-learn
Model
RandomForest / XGBoost
Database
PostgreSQL
Visualization
Heatmap + markers
Deployment
Vercel + Railway/Render

Important rule:
Map tiles can be used only for visualization.
No external map/traffic/geocoding data should be used for training or scoring.

To be safest, your model should depend only on:
latitude
longitude
created_datetime
vehicle_type
violation_type
police_station
junction_name
validation_status


9. Core ML Design
Use a two-layer solution.

Layer 1: Illegal Parking Hotspot Prediction
Objective
Predict whether a zone will become a high-risk illegal parking hotspot in the next time window.
Training unit
one row = one zone + one time window

Example:
zone_id = grid_12.977_77.580
time_window = 2024-02-12 18:00-19:00

Input features
zone_id
latitude_center
longitude_center
hour_of_day
day_of_week
month
is_weekend
police_station
junction_name_present
violations_last_1h
violations_last_3h
violations_last_24h
violations_last_7d
wrong_parking_count
no_parking_count
main_road_parking_count
double_parking_count
near_crossing_count
footpath_parking_count
heavy_vehicle_count
repeat_hotspot_score

Target
Option 1: Regression
violations_next_1h

Option 2: Classification
LOW / MEDIUM / HIGH hotspot risk

For hackathon, use classification because it is easier to present.
Example target creation:
0-2 violations in next hour   → LOW
3-8 violations in next hour   → MEDIUM
9+ violations in next hour    → HIGH

Thresholds should be adjusted based on data distribution.
Model
Use:
RandomForestClassifier

or:
XGBoostClassifier

For fastest prototype:
RandomForestClassifier


Layer 2: Congestion-Impact Proxy Scoring
Since actual congestion data is unavailable, build a proxy score.
Name it clearly:
Parking Congestion Impact Potential Score

This estimates how likely illegal parking at a zone is to obstruct traffic.
Inputs
predicted_hotspot_risk
violation severity
vehicle type severity
junction proximity
repeat hotspot frequency
time of day
location density

Output
{
  "impact_score": 84,
  "impact_level": "HIGH",
  "reason": [
    "High repeat violation density",
    "Parking in main road detected",
    "Near junction",
    "Heavy vehicle involvement"
  ]
}


10. Severity Scoring Logic
Violation severity weights
Violation type
Weight
Double parking
30
Parking in main road
25
Parking near road crossing
25
Parking near traffic light/zebra crossing
25
Parking near bus stop/school/hospital
20
Parking opposite another parked vehicle
20
No parking
15
Parking on footpath
12
Wrong parking
10
Defective number plate
3


Vehicle severity weights
Vehicle type
Weight
Bus / Private Bus / Tourist Bus / School Vehicle
25
HGV / Lorry / Tanker
25
LGV / Van / Tempo / Maxi-cab
18
Car / Jeep
12
Passenger Auto / Goods Auto
10
Scooter / Motorcycle / Moped
5


Junction weight
No Junction        → 0
Known Junction     → 20
High-repeat junction → 30

Reason:
Parking near intersections is more likely to reduce turning capacity and create bottlenecks.


Final Priority Score
priority_score =
  0.35 * hotspot_prediction_score
+ 0.30 * violation_severity_score
+ 0.20 * repeat_hotspot_score
+ 0.10 * junction_score
+ 0.05 * heavy_vehicle_score

Priority levels:
0-40     LOW
41-70    MEDIUM
71-85    HIGH
86-100   CRITICAL


11. Final Model Input and Output
Model Input Example
{
  "zone_id": "zone_12.977_77.580",
  "hour": 18,
  "day_of_week": "Friday",
  "is_weekend": false,
  "police_station": "Upparpet",
  "junction_name": "BTP044 - Sagar Theatre Junction",
  "violations_last_1h": 7,
  "violations_last_24h": 43,
  "violations_last_7d": 188,
  "wrong_parking_count": 21,
  "no_parking_count": 14,
  "main_road_parking_count": 6,
  "double_parking_count": 2,
  "heavy_vehicle_count": 3,
  "repeat_hotspot_score": 82
}

Final Output Example
{
  "zone_id": "zone_12.977_77.580",
  "latitude": 12.977,
  "longitude": 77.580,
  "police_station": "Upparpet",
  "junction_name": "BTP044 - Sagar Theatre Junction",
  "predicted_hotspot_risk": "HIGH",
  "expected_violations_next_hour": 12,
  "parking_congestion_impact_score": 84,
  "priority_score": 88,
  "priority_level": "CRITICAL",
  "recommended_action": "Deploy towing/enforcement team",
  "reasons": [
    "High repeat illegal parking density",
    "Main road parking violations present",
    "Located near a junction",
    "Heavy vehicle parking detected"
  ]
}


12. Database Design
Table: violations
id
latitude
longitude
location
vehicle_type
violation_type
created_datetime
police_station
junction_name
validation_status
zone_id


Table: zone_time_features
id
zone_id
time_window_start
time_window_end
latitude_center
longitude_center
police_station
junction_name
total_violations
wrong_parking_count
no_parking_count
main_road_parking_count
double_parking_count
near_crossing_count
footpath_parking_count
heavy_vehicle_count
repeat_hotspot_score


Table: zone_predictions
id
zone_id
prediction_time
predicted_hotspot_risk
expected_violations
impact_score
priority_score
priority_level
recommended_action
reasons_json
created_at


13. Backend API Design
1. Get hotspot map
GET /api/hotspots?date=2024-03-10&hour=18&police_station=Upparpet

Response:
{
  "hotspots": [
    {
      "zone_id": "zone_12.977_77.580",
      "lat": 12.977,
      "lng": 77.580,
      "priority_score": 88,
      "priority_level": "CRITICAL",
      "impact_score": 84,
      "expected_violations": 12
    }
  ]
}


2. Get enforcement ranking
GET /api/enforcement/ranking?date=2024-03-10&hour=18

Response:
{
  "rankings": [
    {
      "rank": 1,
      "zone_id": "zone_12.977_77.580",
      "police_station": "Upparpet",
      "junction_name": "BTP044 - Sagar Theatre Junction",
      "priority_score": 88,
      "priority_level": "CRITICAL",
      "recommended_action": "Deploy towing team"
    }
  ]
}


3. Get zone details
GET /api/zones/{zone_id}/insights

Response:
{
  "zone_id": "zone_12.977_77.580",
  "total_violations": 432,
  "repeat_hotspot_score": 91,
  "top_violation_types": [
    "WRONG PARKING",
    "NO PARKING",
    "PARKING IN A MAIN ROAD"
  ],
  "top_vehicle_types": [
    "CAR",
    "SCOOTER",
    "MAXI-CAB"
  ],
  "impact_score": 84,
  "reasons": [
    "Repeat hotspot",
    "Main-road parking",
    "Near junction"
  ]
}


4. Simulation API
POST /api/simulate

Request:
{
  "zone_id": "zone_12.977_77.580",
  "violation_reduction_percent": 50
}

Response:
{
  "current_priority_score": 88,
  "simulated_priority_score": 61,
  "priority_change": "CRITICAL → MEDIUM",
  "estimated_impact_reduction": 27
}

Important wording:
This is a priority-score simulation, not actual traffic-speed simulation.


14. Frontend Dashboard
Page 1: City Hotspot Map
Show:
Map of Bangalore using lat/lon points
Heatmap of illegal parking density
Color-coded priority zones
Date/hour filters
Police station filter
Violation type filter

Marker colors:
Green  → LOW
Yellow → MEDIUM
Orange → HIGH
Red    → CRITICAL


Page 2: Enforcement Ranking
Table:
Rank
Police Station
Junction
Expected Violations
Impact
Priority
1
Upparpet
Sagar Theatre Junction
12
84
CRITICAL
2
City Market
KR Market Junction
10
79
HIGH


Page 3: Zone Detail
Show:
zone location
total historical violations
hour-wise pattern
day-wise pattern
top violation types
top vehicle types
repeat hotspot score
impact score
recommended action


Page 4: Simulation
Slider:
Reduce illegal parking by 0% to 100%

Output:
priority score reduction
impact score reduction
recommended enforcement benefit


15. Working Prototype: Should You Build One?
Yes.
For Flipkart Gridlock, you should present a working prototype, not just slides.
Minimum demo flow:
1. Open dashboard
2. Select police station/date/hour
3. View illegal parking hotspot heatmap
4. Click a hotspot
5. See expected violations, impact score, priority score, and reasons
6. Open enforcement ranking
7. Run simulation: “What if violations reduce by 50%?”
8. Show priority score improvement

That is enough for a strong prototype.

16. Implementation Plan
Phase 1: Data Cleaning
Tasks:
unzip CSV
load into Python
parse created_datetime
extract hour/day/month/weekend
parse violation_type JSON-like array
parse offence_code
remove invalid coordinates
normalize vehicle_type
normalize police_station
normalize junction_name
handle missing validation_status

Do not blindly remove all missing validation rows because many rows have missing validation status.
Better:
approved      → confidence 1.0
created/null  → confidence 0.7
processing    → confidence 0.6
rejected      → confidence 0.2 or exclude
duplicate     → exclude


Phase 2: Zone Creation
Use either:
Option A: Simple lat/lon grid
Option B: H3 hexagon indexing

For hackathon, use simple grid:
zone_lat = round(latitude, 3)
zone_lon = round(longitude, 3)
zone_id = zone_lat + "_" + zone_lon

This gives approximate local zones.
No external dataset required.

Phase 2.1: Principal Component Analysis

Phase 3: Feature Engineering
Aggregate by:
zone_id + hour

or:
zone_id + date + hour

Create features:
violations_last_1h
violations_last_3h
violations_last_24h
violations_last_7d
wrong_parking_count
no_parking_count
main_road_parking_count
double_parking_count
near_crossing_count
near_signal_count
footpath_count
heavy_vehicle_count
junction_present
repeat_hotspot_score


Phase 4: Label Creation
For every zone and hour:
target = violations in next hour

Example:
Features from 5 PM to 6 PM
Target = violations from 6 PM to 7 PM

Convert to class:
LOW     → 0 to 2 violations
MEDIUM  → 3 to 8 violations
HIGH    → 9+ violations

Use percentile-based thresholds if class imbalance is high.

Phase 5: Model Training
XGBoostClassifier

Evaluation:
accuracy
precision/recall for HIGH-risk class
F1 score
confusion matrix
top-K hotspot recall

Most important metric for demo:
How many actual high-risk zones appear in the model’s top 10/top 20 predictions?

Because enforcement teams care about top-priority locations.

Phase 6: Scoring Layer
For every predicted hotspot:
calculate severity score
calculate repeat score
calculate junction score
calculate vehicle obstruction score
calculate final priority score

Return:
priority_level
recommended_action
reasons


Phase 7: Backend
Use Go backend.
Responsibilities:
serve hotspot data
serve enforcement ranking
serve zone insights
serve simulation results
read predictions from database
optionally call Python ML service

For stable demo:
Precompute predictions offline.
Store predictions in PostgreSQL.
Go backend only serves results.


Phase 8: Frontend
Build:
Next.js dashboard
Leaflet map
heatmap layer
ranking table
detail side panel
simulation slider


17. Prototype Deployment Strategy
Recommended:
Frontend: Vercel
Backend: Railway
Database: Supabase Postgres
ML training: offline locally
Predictions: precomputed and inserted into DB

Avoid live model training during demo.
Safer:
Train model offline
Generate predictions.csv
Load predictions into database
Backend serves prediction results


18. Edge Cases
Dataset edge cases
Edge case
Handling
Missing location text
Use lat/lon and police station
Missing latitude/longitude
Drop row
Invalid coordinates
Drop row
Missing police station
Use a python script to find out nearest police station to the provided address
Missing junction name
Mark as No Junction
No Junction value
junction_score = 0
Multiple violation types
Parse and count each one
Invalid violation JSON string
Treat as single raw violation
Duplicate records
Remove if same vehicle/time/location
Rejected violations
Exclude or give low confidence
Duplicate validation status
Exclude
Missing validation status
Keep with lower confidence
Outlier coordinates
Remove if outside Bangalore bounding box
Vehicle type missing
Mark as UNKNOWN_VEHICLE, or label as most present vehicle type






ML edge cases
Edge case
Handling
Class imbalance
Use class weights or percentile thresholds
Zone with little history
Lower model confidence
New zone unseen in training
Use nearest police-station average
Expected violations negative
Clamp to 0
Probability too close
Mark confidence as medium/low
Overfitting
Use time-based train/test split
Data leakage
Do not use future violation counts
Sparse night data
Merge low-volume hours if needed
Repeated exact coordinates
Cap extreme density influence
Model predicts all low risk
Adjust class thresholds


Scoring edge cases
Edge case
Handling
High violations but mostly scooters
Lower vehicle obstruction score
Low count but double parking near junction
Increase impact score
High count but no junction
Medium impact unless repeat hotspot
Defective number plate only
Low congestion impact
Heavy vehicle parked on main road
High impact even with low count
Police station with high patrol activity
Normalize by station-level average
Rejected records concentrated in one zone
Lower confidence


Backend edge cases
Edge case
Handling
Invalid date/hour
Return 400
Unknown police station
Return supported values
No hotspots found
Return empty list with message
Database down
Serve cached JSON fallback
Large heatmap response
Return top N zones
Slow API
Cache ranking response
Missing prediction
Recompute using fallback score


Frontend edge cases
Edge case
Handling
No data for selected filter
Show empty state
Marker overlap
Use clustering/heatmap
Missing coordinates
Show in table, not map
API failure
Show retry button
Mobile layout
Use collapsible panel
Many zones
Render heatmap, not all markers


19. Final Demo Script
Use this in presentation:
Illegal parking is not equally harmful everywhere. A wrong-parked scooter on a side lane may not matter much, but double parking by a heavy vehicle near a junction can choke movement.

Our system uses only the provided police violation dataset to identify repeated illegal parking hotspots. It learns historical zone-time patterns and predicts where violations are likely to occur next. Since the dataset does not contain actual speed or traffic volume, we estimate congestion-impact potential using violation severity, vehicle obstruction level, junction proximity, and repeat hotspot behavior.

The final output is a ranked enforcement list showing where traffic police should act first, along with explanations for every recommendation.


20. Final Deliverables
1. Cleaned dataset
2. Zone-wise aggregated dataset
3. Trained hotspot prediction model
4. Congestion-impact proxy scoring logic
5. Go backend APIs
6. Next.js dashboard
7. Map heatmap
8. Enforcement ranking table
9. Zone insight panel
10. Simulation feature
11. README
12. Demo video/slides


21. Best Final Architecture Choice
Use this:
Two-layer AI decision-support system

Layer 1: Hotspot Prediction
Predicts:
Where and when illegal parking is likely to occur.

Layer 2: Impact and Priority Scoring
Estimates:
How severe the hotspot is likely to be for traffic movement.

Final output:
Ranked enforcement zones with explainable reasons.

This is the safest, most technically correct, and hackathon-ready approach when only the police violation dataset is allowed.



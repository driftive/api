ALTER TABLE drift_analysis_project
    ADD COLUMN resources_added     INT,
    ADD COLUMN resources_changed   INT,
    ADD COLUMN resources_destroyed INT;

UPDATE drift_analysis_project
SET resources_added     = (regexp_match(plan_output, 'Plan:\s+(\d+) to add,\s+(\d+) to change,\s+(\d+) to destroy'))[1]::int,
    resources_changed   = (regexp_match(plan_output, 'Plan:\s+(\d+) to add,\s+(\d+) to change,\s+(\d+) to destroy'))[2]::int,
    resources_destroyed = (regexp_match(plan_output, 'Plan:\s+(\d+) to add,\s+(\d+) to change,\s+(\d+) to destroy'))[3]::int
WHERE plan_output ~ 'Plan:\s+\d+ to add,\s+\d+ to change,\s+\d+ to destroy';

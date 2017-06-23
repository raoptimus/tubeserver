db.PushTask.find({"Options.ElapseDaysLastActive": {"$exists": true}}).forEach(function (task) {
    task.Options = {
        "ElapseDaysLastActiveFrom": task.Options.ElapseDaysLastActive,
        "ElapseDaysLastActiveTo": 0
    };
    db.PushTask.save(task);
    print(task._id);
});